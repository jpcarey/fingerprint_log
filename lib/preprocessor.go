package lib

import (
	"fmt"
	"regexp"
	"strings"
)

var (
	printDebug = true
)

var (
	ipv4 = regexp.MustCompile(`(?m)((?:(?:25[0-5]|2[0-4][0-9]|[0-1]?[0-9]{1,2})` +
		`\.(?:25[0-5]|2[0-4][0-9]|[0-1]?[0-9]{1,2})` +
		`\.(?:25[0-5]|2[0-4][0-9]|[0-1]?[0-9]{1,2})` +
		`\.(?:25[0-5]|2[0-4][0-9]|[0-1]?[0-9]{1,2}))(?::\d+)?)`,
	)
	uid           = regexp.MustCompile(`(?m)[0-9a-z]{32}`)
	node          = regexp.MustCompile(`node \[?([A-Za-z0-9-_]{22})\]?`)
	node2         = regexp.MustCompile(`((?:{.*?}){6})`)
	nodeName      = regexp.MustCompile(`^(?:\[.*?\]){3} \[(?P<NodeName>.*?)\]`)
	hexString     = regexp.MustCompile(`0x(?:[A-Fa-f0-9]{8})`)
	indexAndShard = regexp.MustCompile(`\[([^ "\*\\<|,>/?]+)\]\[(\d+)\]`)
)

// StripIndetifyingData removes nodename, nodeid, ipaddress, and hex strings
func StripIndetifyingData(line string) string {

	if printDebug {
		fmt.Println(line)
	}

	match := nodeName.FindStringSubmatch(line)
	result := make(map[string]string)
	for i, name := range nodeName.SubexpNames() {
		if i != 0 && name != "" {
			result[name] = match[i]
		}
	}
	line = strings.Replace(line, result["NodeName"], "node_name", -1)

	// fmt.Printf("by name: %s\n", result["NodeName"])

	// name := regexp.MustCompile(result["NodeName"])
	// line = name.ReplaceAllString(line, "")

	line = ipv4.ReplaceAllString(line, "")
	// line = uid.ReplaceAllString(line, "")
	// line = node.ReplaceAllString(line, "")
	line = node.ReplaceAllString(line, "node [node_id]")
	line = hexString.ReplaceAllString(line, "0x00000000")
	line = node2.ReplaceAllString(line, "")
	line = indexAndShard.ReplaceAllString(line, "[index][shard]")

	if printDebug {
		fmt.Println(line)
	}

	return line
}
