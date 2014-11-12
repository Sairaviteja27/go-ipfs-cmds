package cli

import (
	"fmt"
	"io"
	"strings"
	"text/template"

	cmds "github.com/jbenet/go-ipfs/commands"
)

const (
	requiredArg = "<%v>"
	optionalArg = "[<%v>]"
	variadicArg = "%v..."
	optionFlag  = "-%v"
	optionType  = "(%v)"

	whitespace = "\r\n\t "

	indentStr = "    "
)

type helpFields struct {
	Indent      string
	Path        string
	ArgUsage    string
	Tagline     string
	Arguments   string
	Options     string
	Subcommands string
	Description string
}

const usageFormat = "{{.Path}}{{if .ArgUsage}} {{.ArgUsage}}{{end}} - {{.Tagline}}"

const longHelpFormat = `
{{.Indent}}{{template "usage" .}}

{{if .Arguments}}ARGUMENTS:

{{.Indent}}{{.Arguments}}

{{end}}{{if .Options}}OPTIONS:

{{.Indent}}{{.Options}}

{{end}}{{if .Subcommands}}SUBCOMMANDS:

{{.Indent}}{{.Subcommands}}

{{.Indent}}Use '{{.Path}} <subcmd> --help' for more information about each command.

{{end}}{{if .Description}}DESCRIPTION:

{{.Indent}}{{.Description}}

{{end}}
`

var longHelpTemplate *template.Template
var usageTemplate *template.Template

func init() {
	tmpl, err := template.New("usage").Parse(usageFormat)
	if err != nil {
		panic(err)
	}
	usageTemplate = tmpl

	tmpl, err = usageTemplate.New("longHelp").Parse(longHelpFormat)
	if err != nil {
		panic(err)
	}
	longHelpTemplate = tmpl
}

// LongHelp returns a formatted CLI helptext string, generated for the given command
func LongHelp(rootName string, root *cmds.Command, path []string, out io.Writer) error {
	cmd, err := root.Get(path)
	if err != nil {
		return err
	}

	pathStr := rootName
	if len(path) > 0 {
		pathStr += " " + strings.Join(path, " ")
	}

	fields := helpFields{
		Indent:      indentStr,
		Path:        pathStr,
		ArgUsage:    usageText(cmd),
		Tagline:     cmd.Description,
		Arguments:   cmd.ArgumentHelp,
		Options:     cmd.OptionHelp,
		Subcommands: cmd.SubcommandHelp,
		Description: cmd.Help,
	}

	// autogen fields that are empty
	if len(cmd.ArgumentHelp) == 0 {
		fields.Arguments = strings.Join(argumentText(cmd), "\n")
	}
	if len(cmd.OptionHelp) == 0 {
		fields.Options = strings.Join(optionText(cmd), "\n")
	}
	if len(cmd.SubcommandHelp) == 0 {
		fields.Subcommands = strings.Join(subcommandText(cmd, rootName, path), "\n")
	}

	fields.Arguments = indentString(fields.Arguments, indentStr)
	fields.Options = indentString(fields.Options, indentStr)
	fields.Subcommands = indentString(fields.Subcommands, indentStr)
	fields.Description = indentString(fields.Description, indentStr)

	return longHelpTemplate.Execute(out, fields)
}

func argumentText(cmd *cmds.Command) []string {
	lines := make([]string, len(cmd.Arguments))

	for i, arg := range cmd.Arguments {
		lines[i] = argUsageText(arg)
		lines[i] += "\n" + arg.Description
		lines[i] = indentString(lines[i], "    ") + "\n"
	}

	return lines
}

func optionText(cmd ...*cmds.Command) []string {
	// get a slice of the options we want to list out
	options := make([]cmds.Option, 0)
	for _, c := range cmd {
		for _, opt := range c.Options {
			options = append(options, opt)
		}
	}

	// add option names to output (with each name aligned)
	lines := make([]string, 0)
	j := 0
	for {
		done := true
		i := 0
		for _, opt := range options {
			if len(lines) < i+1 {
				lines = append(lines, "")
			}
			if len(opt.Names) >= j+1 {
				lines[i] += fmt.Sprintf(optionFlag, opt.Names[j])
			}
			if len(opt.Names) > j+1 {
				lines[i] += ", "
				done = false
			}

			i++
		}

		if done {
			break
		}

		lines = align(lines)
		j++
	}

	// add option types to output
	for i, opt := range options {
		lines[i] += " " + fmt.Sprintf(optionType, opt.Type)
	}
	lines = align(lines)

	// add option descriptions to output
	for i, opt := range options {
		lines[i] += "\n" + opt.Description
		lines[i] = indentString(lines[i], "    ") + "\n"
	}

	return lines
}

func subcommandText(cmd *cmds.Command, rootName string, path []string) []string {
	prefix := fmt.Sprintf("%v %v", rootName, strings.Join(path, " "))
	if len(path) > 0 {
		prefix += " "
	}
	lines := make([]string, len(cmd.Subcommands))

	i := 0
	for name, sub := range cmd.Subcommands {
		usage := usageText(sub)
		lines[i] = fmt.Sprintf("%v%v %v", prefix, name, usage)
		lines[i] += fmt.Sprintf("\n%v", sub.Description)
		lines[i] = indentString(lines[i], "    ") + "\n"
		i++
	}

	return lines
}

func usageText(cmd *cmds.Command) string {
	s := ""
	for i, arg := range cmd.Arguments {
		if i != 0 {
			s += " "
		}
		s += argUsageText(arg)
	}

	return s
}

func argUsageText(arg cmds.Argument) string {
	s := arg.Name

	if arg.Required {
		s = fmt.Sprintf(requiredArg, s)
	} else {
		s = fmt.Sprintf(optionalArg, s)
	}

	if arg.Variadic {
		s = fmt.Sprintf(variadicArg, s)
	}

	return s
}

func align(lines []string) []string {
	longest := 0
	for _, line := range lines {
		length := len(line)
		if length > longest {
			longest = length
		}
	}

	for i, line := range lines {
		length := len(line)
		if length > 0 {
			lines[i] += strings.Repeat(" ", longest-length)
		}
	}

	return lines
}

func indent(lines []string, prefix string) []string {
	for i, line := range lines {
		lines[i] = prefix + indentString(line, prefix)
	}
	return lines
}

func indentString(line string, prefix string) string {
	return strings.Replace(line, "\n", "\n"+prefix, -1)
}
