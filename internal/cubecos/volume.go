package cubecos

import (
	"encoding/json"
	"os/exec"
)

type MoveBlocker struct {
	Code   string `json:"code"`
	Reason string `json:"reason"`
}

type MovePreflight struct {
	OK     bool   `json:"ok"`
	Code   string `json:"code"`
	Reason string `json:"reason"`
	// Blockers carries every reason the move cannot proceed; Code/Reason
	// repeat the first so single-line callers stay simple.
	Blockers   []MoveBlocker `json:"blockers"`
	SrcType    string        `json:"srcType"`
	DstType    string        `json:"dstType"`
	SizeGb     int           `json:"sizeGb"`
	AttachedTo string        `json:"attachedTo"`
}

// rawMovePreflight mirrors cinder_move_preflight's stdout, which uses
// snake_case keys. MovePreflight uses camelCase to match this API's own
// response convention, so the two are kept separate and converted.
type rawMovePreflight struct {
	OK         bool          `json:"ok"`
	Code       string        `json:"code"`
	Reason     string        `json:"reason"`
	Blockers   []MoveBlocker `json:"blockers"`
	SrcType    string        `json:"src_type"`
	DstType    string        `json:"dst_type"`
	SizeGb     int           `json:"size_gb"`
	AttachedTo string        `json:"attached_to"`
}

// MovePreflightStatus maps a preflight code to the HTTP status the API returns.
// A bad request names something that does not exist or is a no-op; a conflict is
// a real volume/instance state the caller must change first.
func MovePreflightStatus(code string) int {
	switch code {
	case "OK":
		return 200
	case "E_NO_SUCH_TYPE", "E_SAME_TYPE":
		return 400
	default:
		return 409
	}
}

// RunMovePreflight asks hex_sdk whether volumeID may move to destType.
// cinder_move_preflight prints JSON on both success and refusal, exiting
// non-zero on refusal, so non-empty stdout is authoritative over the exit
// status.
func RunMovePreflight(volumeID, destType string) (*MovePreflight, error) {
	out, err := exec.Command("hex_sdk", "cinder_move_preflight", volumeID, destType).Output()
	if len(out) == 0 && err != nil {
		return nil, err
	}

	var raw rawMovePreflight
	if jerr := json.Unmarshal(out, &raw); jerr != nil {
		return nil, jerr
	}

	pf := MovePreflight(raw)
	return &pf, nil
}

// MoveDispatch is cinder_move_volume's result: the same preflight fields plus
// whether the retype was actually dispatched.
type MoveDispatch struct {
	OK     bool   `json:"ok"`
	Code   string `json:"code"`
	Reason string `json:"reason"`
	// Blockers carries every reason the move cannot proceed; Code/Reason
	// repeat the first so single-line callers stay simple.
	Blockers   []MoveBlocker `json:"blockers"`
	SrcType    string        `json:"srcType"`
	DstType    string        `json:"dstType"`
	SizeGb     int           `json:"sizeGb"`
	AttachedTo string        `json:"attachedTo"`
	Dispatched bool          `json:"dispatched"`
}

// rawMoveDispatch mirrors cinder_move_volume's stdout (snake_case keys), same
// as rawMovePreflight, plus the dispatched flag.
type rawMoveDispatch struct {
	OK         bool          `json:"ok"`
	Code       string        `json:"code"`
	Reason     string        `json:"reason"`
	Blockers   []MoveBlocker `json:"blockers"`
	SrcType    string        `json:"src_type"`
	DstType    string        `json:"dst_type"`
	SizeGb     int           `json:"size_gb"`
	AttachedTo string        `json:"attached_to"`
	Dispatched bool          `json:"dispatched"`
}

// MoveDispatchStatus maps a cinder_move_volume result code to the HTTP status
// the move endpoint returns. A refusal maps exactly as the preflight does; a
// dispatch failure (E_DISPATCH_FAILED) is a server-side failure, not a client
// mistake; a successful dispatch is 202 Accepted, not 200.
func MoveDispatchStatus(code string) int {
	switch code {
	case "E_DISPATCH_FAILED":
		return 500
	case "OK":
		return 202
	default:
		return MovePreflightStatus(code)
	}
}

// RunMoveVolume asks hex_sdk to move volumeID to destType. cinder_move_volume
// runs the preflight itself and, if it passes, dispatches the retype. It
// prints JSON on refusal, dispatch failure, and success alike, exiting
// non-zero on the first two, so non-empty stdout is authoritative over the
// exit status, exactly as with the preflight.
func RunMoveVolume(volumeID, destType string) (*MoveDispatch, error) {
	out, err := exec.Command("hex_sdk", "cinder_move_volume", volumeID, destType).Output()
	if len(out) == 0 && err != nil {
		return nil, err
	}

	var raw rawMoveDispatch
	if jerr := json.Unmarshal(out, &raw); jerr != nil {
		return nil, jerr
	}

	md := MoveDispatch(raw)
	return &md, nil
}
