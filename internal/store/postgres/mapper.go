package postgres

import (
	"encoding/json"
	"fmt"

	domaingame "github.com/MehrshadFb/xo-grpc/internal/domain/game"
)

func markToString(mark domaingame.Mark) string {
	switch mark {
	case domaingame.MarkX:
		return "X"
	case domaingame.MarkO:
		return "O"
	case domaingame.MarkEmpty:
		return "EMPTY"
	default:
		return "UNSPECIFIED"
	}
}

func stringToMark(value string) domaingame.Mark {
	switch value {
	case "X":
		return domaingame.MarkX
	case "O":
		return domaingame.MarkO
	case "EMPTY":
		return domaingame.MarkEmpty
	default:
		return domaingame.MarkUnspecified
	}
}

func statusToString(status domaingame.GameStatus) string {
	switch status {
	case domaingame.StatusWaiting:
		return "WAITING"
	case domaingame.StatusInProgress:
		return "IN_PROGRESS"
	case domaingame.StatusFinished:
		return "FINISHED"
	case domaingame.StatusAborted:
		return "ABORTED"
	default:
		return "UNSPECIFIED"
	}
}

func stringToStatus(value string) domaingame.GameStatus {
	switch value {
	case "WAITING":
		return domaingame.StatusWaiting
	case "IN_PROGRESS":
		return domaingame.StatusInProgress
	case "FINISHED":
		return domaingame.StatusFinished
	case "ABORTED":
		return domaingame.StatusAborted
	default:
		return domaingame.StatusUnspecified
	}
}

func boardToJSON(board [9]domaingame.Mark) ([]byte, error) {
	values := make([]string, len(board))

	for i, mark := range board {
		values[i] = markToString(mark)
	}

	return json.Marshal(values)
}

func jsonToBoard(data []byte) ([9]domaingame.Mark, error) {
	var board [9]domaingame.Mark

	var values []string
	if err := json.Unmarshal(data, &values); err != nil {
		return board, fmt.Errorf("unmarshal board: %w", err)
	}

	if len(values) != 9 {
		return board, fmt.Errorf("invalid board length: %d", len(values))
	}

	for i, value := range values {
		board[i] = stringToMark(value)
	}

	return board, nil
}
