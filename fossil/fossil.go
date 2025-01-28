package fossil

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/NethermindEth/juno/core/felt"
	"github.com/NethermindEth/starknet.go/curve"
)

type JobRequestParams struct {
	Twap         [2]uint64 `json:"twap"`
	Volatility   [2]uint64 `json:"volatility"`
	ReservePrice [2]uint64 `json:"reserve_price"`
}

type ClientInfo struct {
	ClientAddress string `json:"client_address"`
	VaultAddress  string `json:"vault_address"`
	Timestamp     uint64 `json:"timestamp"`
}

type JobRequest struct {
	Identifiers []string         `json:"identifiers"`
	Params      JobRequestParams `json:"params"`
	ClientInfo  ClientInfo       `json:"client_info"`
}

type StatusData struct {
	Status string `json:"status,omitempty"`
	Error  string `json:"error,omitempty"`
}

type FossilResponse struct {
	Status string      `json:"status"`
	Error  string      `json:"error,omitempty"`
	Data   interface{} `json:"data,omitempty"`
}

func createJobRequestParams(targetTimestamp, roundDuration uint64) JobRequestParams {
	return JobRequestParams{
		// TWAP duration is 1 x round duration
		Twap: [2]uint64{targetTimestamp - roundDuration, targetTimestamp},
		// Volatility duration is 3 x round duration
		Volatility: [2]uint64{targetTimestamp - 3*roundDuration, targetTimestamp},
		// Reserve price duration is 3 x round duration
		ReservePrice: [2]uint64{targetTimestamp - 3*roundDuration, targetTimestamp},
	}
}

func CreateJobRequest(targetTimestamp, roundDuration uint64, clientAddress, vaultAddress string) ([]byte, error) {
	if targetTimestamp == 0 || roundDuration == 0 || clientAddress == "" || vaultAddress == "" {
		return nil, fmt.Errorf("missing required parameters")
	}

	request := JobRequest{
		Identifiers: []string{"PITCH_LAKE_V1"},
		Params:      createJobRequestParams(targetTimestamp, roundDuration),
		ClientInfo: ClientInfo{
			ClientAddress: clientAddress,
			VaultAddress:  vaultAddress,
			Timestamp:     targetTimestamp,
		},
	}

	return json.Marshal(request)
}

func MakeFossilRequest(targetTimestamp, roundDuration uint64, clientAddress, vaultAddress string) (*FossilResponse, error) {
	fossilAPIURL := os.Getenv("FOSSIL_API_URL")
	fossilAPIKey := os.Getenv("FOSSIL_API_KEY")

	if fossilAPIURL == "" || fossilAPIKey == "" {
		return nil, fmt.Errorf("missing required environment variables: FOSSIL_API_URL or FOSSIL_API_KEY")
	}

	requestBody, err := CreateJobRequest(targetTimestamp, roundDuration, clientAddress, vaultAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	client := &http.Client{
		Timeout: time.Second * 10,
	}

	req, err := http.NewRequest("POST", fossilAPIURL+"/pricing_data", strings.NewReader(string(requestBody)))
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", fossilAPIKey)

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return &FossilResponse{
			Error: fmt.Sprintf("Fossil request failed with status %d: %s", resp.StatusCode, string(body)),
		}, nil
	}

	var fossilResp FossilResponse
	if err := json.Unmarshal(body, &fossilResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &fossilResp, nil
}

func CreateJobID(targetTimestamp, roundDuration uint64) string {
	if targetTimestamp == 0 || roundDuration == 0 {
		return ""
	}

	identifiers := []string{"PITCH_LAKE_V1"}
	params := createJobRequestParams(targetTimestamp, roundDuration)

	// Convert all inputs to felt.Felt
	identifierStr := strings.Join(identifiers, "")
	identifierFelt := new(felt.Felt).SetBytes([]byte(identifierStr))

	// Convert timestamps to felts
	twap0Felt := new(felt.Felt).SetUint64(uint64(params.Twap[0]))
	twap1Felt := new(felt.Felt).SetUint64(uint64(params.Twap[1]))
	vol0Felt := new(felt.Felt).SetUint64(uint64(params.Volatility[0]))
	vol1Felt := new(felt.Felt).SetUint64(uint64(params.Volatility[1]))
	reserve0Felt := new(felt.Felt).SetUint64(uint64(params.ReservePrice[0]))
	reserve1Felt := new(felt.Felt).SetUint64(uint64(params.ReservePrice[1]))

	// Use PoseidonArray for hashing all values
	hash := curve.PoseidonArray(
		identifierFelt,
		twap0Felt,
		twap1Felt,
		vol0Felt,
		vol1Felt,
		reserve0Felt,
		reserve1Felt,
	)

	return hash.String()
}

func GetFossilStatus(targetTimestamp, roundDuration uint64) (*StatusData, error) {
	fossilAPIURL := os.Getenv("FOSSIL_API_URL")
	if fossilAPIURL == "" {
		return nil, fmt.Errorf("missing required environment variable: FOSSIL_API_URL")
	}

	jobID := CreateJobID(targetTimestamp, roundDuration)
	if jobID == "" {
		return nil, fmt.Errorf("failed to create job ID")
	}

	client := &http.Client{
		Timeout: time.Second * 10,
	}

	req, err := http.NewRequest("GET", fmt.Sprintf("%s/api/getJobStatus?jobId=%s", fossilAPIURL, jobID), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return &StatusData{
			Error: fmt.Sprintf("fossil status request failed with status %d: %s", resp.StatusCode, string(body)),
		}, nil
	}

	var fossilResp FossilResponse
	if err := json.Unmarshal(body, &fossilResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// Convert the Data field to StatusData
	if fossilResp.Data != nil {
		dataBytes, err := json.Marshal(fossilResp.Data)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal status data: %w", err)
		}

		var statusData StatusData
		if err := json.Unmarshal(dataBytes, &statusData); err != nil {
			return nil, fmt.Errorf("failed to unmarshal status data: %w", err)
		}
		return &statusData, nil
	}

	// If no data, return error from the response
	if fossilResp.Error != "" {
		return &StatusData{
			Error: fossilResp.Error,
		}, nil
	}

	return &StatusData{
		Status: fossilResp.Status,
	}, nil
}
