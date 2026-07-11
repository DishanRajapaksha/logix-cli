package output

import (
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
	"text/tabwriter"
)

var ErrOutput = errors.New("output error")

func ValidateSnapshotFormat(format string) error {
	switch strings.ToLower(format) {
	case "table", "text", "json", "csv":
		return nil
	default:
		return fmt.Errorf("%w: snapshot format must be table, text, json, or csv", ErrOutput)
	}
}

func ValidateStreamFormat(format string) error {
	switch strings.ToLower(format) {
	case "text", "jsonl", "csv":
		return nil
	default:
		return fmt.Errorf("%w: stream format must be text, jsonl, or csv", ErrOutput)
	}
}

func Snapshot(w io.Writer, format string, headers []string, rows [][]string, value any) error {
	if err := ValidateSnapshotFormat(format); err != nil {
		return err
	}
	switch strings.ToLower(format) {
	case "json":
		encoder := json.NewEncoder(w)
		encoder.SetIndent("", "  ")
		if err := encoder.Encode(value); err != nil {
			return fmt.Errorf("%w: %v", ErrOutput, err)
		}
	case "csv":
		writer := csv.NewWriter(w)
		if err := writer.Write(headers); err != nil {
			return fmt.Errorf("%w: %v", ErrOutput, err)
		}
		for _, row := range rows {
			if err := writer.Write(row); err != nil {
				return fmt.Errorf("%w: %v", ErrOutput, err)
			}
		}
		writer.Flush()
		if err := writer.Error(); err != nil {
			return fmt.Errorf("%w: %v", ErrOutput, err)
		}
	case "table":
		writer := tabwriter.NewWriter(w, 0, 4, 2, ' ', 0)
		fmt.Fprintln(writer, strings.Join(headers, "\t"))
		for _, row := range rows {
			fmt.Fprintln(writer, strings.Join(row, "\t"))
		}
		if err := writer.Flush(); err != nil {
			return fmt.Errorf("%w: %v", ErrOutput, err)
		}
	case "text":
		for _, row := range rows {
			fmt.Fprintln(w, strings.Join(row, "\t"))
		}
	}
	return nil
}

type Stream struct {
	w           io.Writer
	format      string
	header      []string
	wroteHeader bool
}

func NewStream(w io.Writer, format string, header []string) (*Stream, error) {
	if err := ValidateStreamFormat(format); err != nil {
		return nil, err
	}
	return &Stream{w: w, format: strings.ToLower(format), header: header}, nil
}

func (s *Stream) Write(row []string, value any) error {
	switch s.format {
	case "jsonl":
		if err := json.NewEncoder(s.w).Encode(value); err != nil {
			return fmt.Errorf("%w: %v", ErrOutput, err)
		}
	case "csv":
		writer := csv.NewWriter(s.w)
		if !s.wroteHeader {
			if err := writer.Write(s.header); err != nil {
				return fmt.Errorf("%w: %v", ErrOutput, err)
			}
			s.wroteHeader = true
		}
		if err := writer.Write(row); err != nil {
			return fmt.Errorf("%w: %v", ErrOutput, err)
		}
		writer.Flush()
		if err := writer.Error(); err != nil {
			return fmt.Errorf("%w: %v", ErrOutput, err)
		}
	case "text":
		fmt.Fprintln(s.w, strings.Join(row, "\t"))
	}
	return nil
}
