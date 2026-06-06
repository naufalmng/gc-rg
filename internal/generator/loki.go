package generator

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
)

type LokiScopeRow struct {
	Job        string
	Instance   string
	LogLines   float64
	ErrorLines float64
	InfoLines  float64
	Status     string
}

type grafanaFrameFile map[string]struct {
	Data struct {
		Results map[string]struct {
			Frames []grafanaFrame `json:"frames"`
		} `json:"results"`
	} `json:"data"`
}

type grafanaFrame struct {
	Schema struct {
		Fields []struct {
			Labels map[string]string `json:"labels"`
		} `json:"fields"`
	} `json:"schema"`
	Data struct {
		Values [][]any `json:"values"`
	} `json:"data"`
}

func LoadLokiScopeRows(path string) ([]LokiScopeRow, error) {
	if path == "" {
		return nil, nil
	}
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, nil
	}
	var data grafanaFrameFile
	bytes, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read Loki scope JSON: %w", err)
	}
	if err := json.Unmarshal(bytes, &data); err != nil {
		return nil, fmt.Errorf("parse Loki scope JSON: %w", err)
	}
	lineRows := frameRows(data, "loki_label_job_instance")
	errorRows := frameRows(data, "loki_error_by_job_instance")
	infoRows := frameRows(data, "loki_info_by_job_instance")
	keys := sortedKeys(lineRows, errorRows, infoRows)
	rows := make([]LokiScopeRow, 0, len(keys))
	for _, key := range keys {
		logLines := lineRows[key]
		errorLines := fallbackValue(errorRows, key)
		infoLines := fallbackValue(infoRows, key)
		rows = append(rows, LokiScopeRow{
			Job:        key.Job,
			Instance:   key.Instance,
			LogLines:   logLines,
			ErrorLines: errorLines,
			InfoLines:  infoLines,
			Status:     logStatus(errorLines, infoLines, logLines),
		})
	}
	return rows, nil
}

type lokiKey struct{ Job, Instance string }

func frameRows(data grafanaFrameFile, queryName string) map[lokiKey]float64 {
	result := map[lokiKey]float64{}
	query, ok := data[queryName]
	if !ok {
		return result
	}
	frames := query.Data.Results["A"].Frames
	for _, frame := range frames {
		labels := map[string]string{}
		for _, field := range frame.Schema.Fields {
			for key, value := range field.Labels {
				labels[key] = value
			}
		}
		values := frame.Data.Values
		if len(values) < 2 || len(values[1]) == 0 {
			continue
		}
		value, ok := asFloat(values[1][len(values[1])-1])
		if !ok {
			continue
		}
		result[lokiKey{Job: labels["job"], Instance: defaultString(labels["instance"], "-")}] = value
	}
	return result
}

func sortedKeys(maps ...map[lokiKey]float64) []lokiKey {
	seen := map[lokiKey]bool{}
	for _, item := range maps {
		for key := range item {
			seen[key] = true
		}
	}
	keys := make([]lokiKey, 0, len(seen))
	for key := range seen {
		keys = append(keys, key)
	}
	sort.Slice(keys, func(i, j int) bool {
		if keys[i].Job == keys[j].Job {
			return keys[i].Instance < keys[j].Instance
		}
		return keys[i].Job < keys[j].Job
	})
	return keys
}

func fallbackValue(values map[lokiKey]float64, key lokiKey) float64 {
	if value, ok := values[key]; ok {
		return value
	}
	return values[lokiKey{Job: key.Job, Instance: "-"}]
}

func logStatus(errorLines, infoLines, logLines float64) string {
	if errorLines > 0 {
		return "Warning"
	}
	if infoLines > 0 || logLines > 0 {
		return "Info"
	}
	return "Normal"
}

func asFloat(value any) (float64, bool) {
	switch typed := value.(type) {
	case float64:
		return typed, true
	case int:
		return float64(typed), true
	default:
		return 0, false
	}
}

func defaultString(value, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
}
