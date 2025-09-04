package common

import "strings"

// List of all subcommands with array parameters to expand.
var arrayParams = map[string][]string{
	"image apply-tags": {"--tags", "-t"},
	"image build":      {"--labels", "-l", "annotations", "-a"},
}

// expandArrayParameters is a workaround for missing pflag ability to parse parameters array separated by spaces.
// We need to process parameters like:
// cli --array v1 v2 v3 --some-arg
// but pflag supports only
// cli --array v1 --array v2 --array v3 --some-arg
// or comma separated values like:
// cli --array v1,v2,v3 --some-arg
// This function expands array parameters, so "--array v1 v2 v3" becomes "--array v1 --array v2 --array v3"
func ExpandArrayParameters(argv []string) []string {
	out := make([]string, 0, len(argv))

	// Determine the command which is run.
	cmdPath := []string{}
	for _, arg := range argv {
		if strings.HasPrefix(arg, "-") {
			break
		}
		cmdPath = append(cmdPath, arg)
	}

	joinedPath := strings.Join(cmdPath, " ")
	multiFlags := map[string]bool{}
	for _, arrayParam := range arrayParams[joinedPath] {
		multiFlags[arrayParam] = true
	}

	for i := 0; i < len(argv); i++ {
		arg := argv[i]

		// Stop processing after "--" sentinel, in case we have positional arguments.
		if arg == "--" {
			out = append(out, argv[i:]...)
			break
		}

		// Handle parameters with = for example: --array1=a1 a2
		if strings.HasPrefix(arg, "--") && strings.Contains(arg, "=") {
			parts := strings.SplitN(arg, "=", 2)
			flag := parts[0]
			if multiFlags[flag] {
				for _, v := range strings.Split(parts[1], ",") {
					if v != "" {
						out = append(out, flag, v)
					}
				}
				continue
			}
			out = append(out, arg)
			continue
		}

		// If this arg is an array, duplicate the arg before each array element.
		if multiFlags[arg] {
			j := i + 1
			for j < len(argv) && argv[j] != "--" && !strings.HasPrefix(argv[j], "-") {
				out = append(out, arg, argv[j])
				j++
			}
			// If no values given, it must be an empty array.
			// Remove the arg, because empty array is not supported by pflag, so it will fail with an error.
			if j == i+1 {
				// do not let pflag fail, omit the array arg
				// out = append(out, arg)
				j++
			}
			i = j - 1
			continue
		}

		out = append(out, arg)
	}
	return out
}
