package output

import (
	"fmt"
	"io"

	shared "github.com/DishanRajapaksha/industrial-cli-kit/output"
)

var ErrOutput = shared.ErrOutput

func ValidateSnapshotFormat(format string) error {
	if err := shared.ValidateSnapshotFormat(format); err != nil {
		return fmt.Errorf("%w: %v", ErrOutput, err)
	}
	return nil
}

func ValidateStreamFormat(format string) error {
	if err := shared.ValidateStreamFormat(format); err != nil {
		return fmt.Errorf("%w: %v", ErrOutput, err)
	}
	return nil
}

func Snapshot(w io.Writer, format string, headers []string, rows [][]string, value any) error {
	switch format {
	case "json":
		return shared.WriteJSON(w, value)
	case "csv":
		return shared.WriteCSV(w, headers, rows)
	case "table":
		return shared.WriteTable(w, headers, rows)
	case "text":
		for _, row := range rows {
			if err := shared.WriteText(w, join(row)); err != nil {
				return err
			}
		}
		return nil
	default:
		return ValidateSnapshotFormat(format)
	}
}

type Stream = shared.Stream

func NewStream(w io.Writer, format string, header []string) (*Stream, error) {
	stream, err := shared.NewStream(w, format, header)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrOutput, err)
	}
	return stream, nil
}
func join(row []string) string {
	if len(row) == 0 {
		return ""
	}
	value := row[0]
	for _, part := range row[1:] {
		value += "\t" + part
	}
	return value
}
