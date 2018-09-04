package lib

import (
	"regexp"
	"strings"
)

var (
	printDebug = false
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
	log4jtilde    = regexp.MustCompile(`\~\[`)
)

func removeTimestamp(line string) string {
	return line[25:]
}

func tolower(line string) string {
	return strings.ToLower(line)
}

// filter removes nodename, nodeid, ipaddress, and hex strings

func filter(s []string) []string {
	// func filter(line string) string {

	// iterate the array of log lines (message)
	for i := range s {
		line := s[i]
		result := make(map[string]string)
		if i == 0 {
			// check if we can get the node name from the first line
			match := nodeName.FindStringSubmatch(line)
			// TODO: need to properly fix this. Not sure the correct way to ensure
			//  the regex matched.
			if len(match) >= 2 {
				for i, name := range nodeName.SubexpNames() {
					if i >= 1 && name != "" {
						result[name] = match[i]
					}
				}
				line = removeTimestamp(line)
			}
		}
		if val, ok := result["NodeName"]; ok {
			line = strings.Replace(line, val, "node_name", -1)
		}
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
		line = log4jtilde.ReplaceAllString(line, "[")

		line = tolower(line)

		s[i] = line
	}
	// line := string(s)
	// if printDebug {
	// 	fmt.Printf("R: %s\n", line)
	// }

	// if printDebug {
	// 	fmt.Printf("P: %s\n", line)
	// }

	return s
}
