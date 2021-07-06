package assert

// log 输出错误
func log(t T, failure string, args []interface{}, opts *Options) {

}

// func Fail(t TestingT, failure string, args ...interface{}) bool {
// 	content := []labeledContent{
// 		{"Error Trace", strings.Join(CallerInfo(), "\n\t\t\t")},
// 		{"Error", failure},
// 	}

// 	// Add test name if the Go version supports it
// 	if n, ok := t.(interface {
// 		Name() string
// 	}); ok {
// 		content = append(content, labeledContent{"Test", n.Name()})
// 	}

// 	message := messageFromMsgAndArgs(args...)
// 	if len(message) > 0 {
// 		content = append(content, labeledContent{"Messages", message})
// 	}

// 	t.Errorf("\n%s", ""+labeledOutput(content...))

// 	return false
// }

// // formatUnequalValues takes two values of arbitrary types and returns string
// // representations appropriate to be presented to the user.
// //
// // If the values are not of like type, the returned strings will be prefixed
// // with the type name, and the value will be enclosed in parenthesis similar
// // to a type conversion in the Go grammar.
// func formatUnequalValues(expected, actual interface{}) (e string, a string) {
// 	if reflect.TypeOf(expected) != reflect.TypeOf(actual) {
// 		return fmt.Sprintf("%T(%s)", expected, truncatingFormat(expected)),
// 			fmt.Sprintf("%T(%s)", actual, truncatingFormat(actual))
// 	}
// 	switch expected.(type) {
// 	case time.Duration:
// 		return fmt.Sprintf("%v", expected), fmt.Sprintf("%v", actual)
// 	}
// 	return truncatingFormat(expected), truncatingFormat(actual)
// }

// // truncatingFormat formats the data and truncates it if it's too long.
// //
// // This helps keep formatted error messages lines from exceeding the
// // bufio.MaxScanTokenSize max line length that the go testing framework imposes.
// func truncatingFormat(data interface{}) string {
// 	value := fmt.Sprintf("%#v", data)
// 	max := bufio.MaxScanTokenSize - 100 // Give us some space the type info too if needed.
// 	if len(value) > max {
// 		value = value[0:max] + "<... truncated>"
// 	}
// 	return value
// }

// type labeledContent struct {
// 	label   string
// 	content string
// }

// // labeledOutput returns a string consisting of the provided labeledContent. Each labeled output is appended in the following manner:
// //
// //   \t{{label}}:{{align_spaces}}\t{{content}}\n
// //
// // The initial carriage return is required to undo/erase any padding added by testing.T.Errorf. The "\t{{label}}:" is for the label.
// // If a label is shorter than the longest label provided, padding spaces are added to make all the labels match in length. Once this
// // alignment is achieved, "\t{{content}}\n" is added for the output.
// //
// // If the content of the labeledOutput contains line breaks, the subsequent lines are aligned so that they start at the same location as the first line.
// func labeledOutput(content ...labeledContent) string {
// 	longestLabel := 0
// 	for _, v := range content {
// 		if len(v.label) > longestLabel {
// 			longestLabel = len(v.label)
// 		}
// 	}
// 	var output string
// 	for _, v := range content {
// 		output += "\t" + v.label + ":" + strings.Repeat(" ", longestLabel-len(v.label)) + "\t" + indentMessageLines(v.content, longestLabel) + "\n"
// 	}
// 	return output
// }

// func messageFromMsgAndArgs(msgAndArgs ...interface{}) string {
// 	if len(msgAndArgs) == 0 || msgAndArgs == nil {
// 		return ""
// 	}
// 	if len(msgAndArgs) == 1 {
// 		msg := msgAndArgs[0]
// 		if msgAsStr, ok := msg.(string); ok {
// 			return msgAsStr
// 		}
// 		return fmt.Sprintf("%+v", msg)
// 	}
// 	if len(msgAndArgs) > 1 {
// 		return fmt.Sprintf(msgAndArgs[0].(string), msgAndArgs[1:]...)
// 	}
// 	return ""
// }

// // Aligns the provided message so that all lines after the first line start at the same location as the first line.
// // Assumes that the first line starts at the correct location (after carriage return, tab, label, spacer and tab).
// // The longestLabelLen parameter specifies the length of the longest label in the output (required becaues this is the
// // basis on which the alignment occurs).
// func indentMessageLines(message string, longestLabelLen int) string {
// 	outBuf := new(bytes.Buffer)

// 	for i, scanner := 0, bufio.NewScanner(strings.NewReader(message)); scanner.Scan(); i++ {
// 		// no need to align first line because it starts at the correct location (after the label)
// 		if i != 0 {
// 			// append alignLen+1 spaces to align with "{{longestLabel}}:" before adding tab
// 			outBuf.WriteString("\n\t" + strings.Repeat(" ", longestLabelLen+1) + "\t")
// 		}
// 		outBuf.WriteString(scanner.Text())
// 	}

// 	return outBuf.String()
// }

// /* CallerInfo is necessary because the assert functions use the testing object
// internally, causing it to print the file:line of the assert method, rather than where
// the problem actually occurred in calling code.*/

// // CallerInfo returns an array of strings containing the file and line number
// // of each stack frame leading from the current test to the assert call that
// // failed.
// func CallerInfo() []string {

// 	var pc uintptr
// 	var ok bool
// 	var file string
// 	var line int
// 	var name string

// 	callers := []string{}
// 	for i := 0; ; i++ {
// 		pc, file, line, ok = runtime.Caller(i)
// 		if !ok {
// 			// The breaks below failed to terminate the loop, and we ran off the
// 			// end of the call stack.
// 			break
// 		}

// 		// This is a huge edge case, but it will panic if this is the case, see #180
// 		if file == "<autogenerated>" {
// 			break
// 		}

// 		f := runtime.FuncForPC(pc)
// 		if f == nil {
// 			break
// 		}
// 		name = f.Name()

// 		// testing.tRunner is the standard library function that calls
// 		// tests. Subtests are called directly by tRunner, without going through
// 		// the Test/Benchmark/Example function that contains the t.Run calls, so
// 		// with subtests we should break when we hit tRunner, without adding it
// 		// to the list of callers.
// 		if name == "testing.tRunner" {
// 			break
// 		}

// 		parts := strings.Split(file, "/")
// 		file = parts[len(parts)-1]
// 		if len(parts) > 1 {
// 			dir := parts[len(parts)-2]
// 			if (dir != "assert" && dir != "mock" && dir != "require") || file == "mock_test.go" {
// 				callers = append(callers, fmt.Sprintf("%s:%d", file, line))
// 			}
// 		}

// 		// Drop the package
// 		segments := strings.Split(name, ".")
// 		name = segments[len(segments)-1]
// 		if isTest(name, "Test") ||
// 			isTest(name, "Benchmark") ||
// 			isTest(name, "Example") {
// 			break
// 		}
// 	}

// 	return callers
// }

// func diff(expected, actual interface{}) string {
// 	return ""
// }
