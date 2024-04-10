package video

import (
	"encoding/json"
	"fmt"

	ffmpeg "github.com/u2takey/ffmpeg-go"
)

func GetTagMetadata(path string) (result map[string]string, err error) {
	jsonProbe, err := ffmpeg.Probe(path)
	if err != nil {
		return nil, err
	}
	mapObject := jsonToMap(jsonProbe)

	result = make(map[string]string)
	if format, ok := mapObject["format"].(map[string]interface{}); ok {
		if tags, ok := format["tags"].(map[string]interface{}); ok {
			for k, v := range tags {
				result[k] = v.(string)
			}
		} else {
			return nil, fmt.Errorf(fmt.Sprintf("no 'tags' found in %s", path))
		}
	} else {
		return nil, fmt.Errorf(fmt.Sprintf("no 'format' found in %s", path))
	}

	return result, nil
}

func SetTagMetadata(inputPath string, tags map[string]string, outputPath string) (err error) {
	command := ffmpeg.Input(inputPath).Output(outputPath, ffmpeg.KwArgs{"c": "copy"}).Compile()
	// multiple -metadata argument are not supports so 
	// they are injected directly into the command instead
	// as a workaround
	command.Args = injectMetadata(command.Args, tags)
	return command.Run()
}

func injectMetadata(parameterList []string, tags map[string]string) []string {
	result := make([]string, 0)

	for _, parameter := range parameterList {
		if parameter == "-c" {
			for key, value := range tags {
				result = append(result, "-metadata", fmt.Sprintf("%s=\"%s\"", key, value))
			}
		}
		result = append(result, parameter)
	}

	return result
}

func jsonToMap(jsonStr string) map[string]interface{} {
	result := make(map[string]interface{})
	json.Unmarshal([]byte(jsonStr), &result)
	return result
}
