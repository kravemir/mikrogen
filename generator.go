package mikrogen

import (
	"fmt"
	"sort"
	"strings"
)

type generator struct {
	Configuration

	builder strings.Builder
}

func (g *generator) generate() string {
	g.generateOldCleanup()

	for _, name := range g.sortedKeys() {
		blocker := g.AccessFilters[name]

		g.generateAddressList(name, blocker)
		g.generateFirewallFilter(name, blocker)
		g.generateToggleScripts(name, blocker)
		g.generateScheduler(name, blocker)
	}

	g.generatePrintInfo()

	return g.builder.String()
}

func (g *generator) sortedKeys() []string {
	keys := []string{}
	for key := range g.AccessFilters {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func (g *generator) generateOldCleanup() {
	g.printSectionf(`Remove old "%s" configuration`, g.IdentifierPrefix)

	g.writeTable([][]string{
		{
			"/system scheduler remove",
			fmt.Sprintf(`[/system scheduler find name~"%s*"]`, g.IdentifierPrefix),
		},
		{
			"/system script remove",
			fmt.Sprintf(`[/system script find name~"%s*"]`, g.IdentifierPrefix),
		},
		{
			"/ip firewall filter remove",
			fmt.Sprintf(`[/ip firewall filter find comment~"%s*"]`, g.IdentifierPrefix),
		},
		{
			"/ip firewall address-list remove",
			fmt.Sprintf(`[/ip firewall address-list find list~"%s*" dynamic=no]`, g.IdentifierPrefix),
		},
	})
}

func (g *generator) generateAddressList(name string, blocker AccessFilter) {
	blockerPrefix := g.IdentifierPrefix + ":" + name

	g.printSectionf("create address-list %s", blockerPrefix)

	g.writeLine("/ip firewall address-list")
	g.writeLine("")

	rows := [][]string{}
	for _, address := range blocker.TargetAddresses {
		rows = append(rows, []string{
			fmt.Sprintf("add address=%s", address),
			fmt.Sprintf("list=%s", blockerPrefix),
		})
	}
	g.writeTable(rows)
}

func (g *generator) generateFirewallFilter(name string, blocker AccessFilter) {
	blockerPrefix := g.IdentifierPrefix + ":" + name

	g.printSectionf("configure firewall filter rules: %s", blockerPrefix)

	g.writeLine("/ip firewall filter")
	g.writeLine("")

	g.writeLine(fmt.Sprintf(
		`add comment="%s" action=reject chain=forward dst-address-list="%s" reject-with=icmp-network-unreachable`,
		blockerPrefix+":DNS",
		blockerPrefix,
	))
	g.writeLine("")

	for _, address := range blocker.TargetTLSHosts {
		g.writeLine(fmt.Sprintf(
			`add comment="%s" action=reject chain=forward protocol=tcp reject-with=icmp-network-unreachable tls-host="%s"`,
			blockerPrefix+":TLS",
			address,
		))
	}

	g.writeLine("")
	g.writeLine(fmt.Sprintf(
		// TODO: hardcoded defconf
		`move destination=([find comment~"defconf*"]->0) numbers=[/ip firewall filter find comment~"%s*"]`,
		blockerPrefix,
	))
}

func (g *generator) generateToggleScripts(name string, _ AccessFilter) {
	blockerPrefix := g.IdentifierPrefix + ":" + name

	g.printSectionf("create scripts to enable / disable filters")

	g.writeLine("/system script")

	g.writeLine(fmt.Sprintf(
		`add name="%s" source="/foreach rule in=[/ip firewall filter find comment~\"%s*\"] do={ /ip firewall filter set \$rule disabled=no }"`,
		blockerPrefix+":Enable",
		blockerPrefix,
	))
	g.writeLine(fmt.Sprintf(
		`add name="%s" source="/foreach rule in=[/ip firewall filter find comment~\"%s*\"] do={ /ip firewall filter set \$rule disabled=yes }"`,
		blockerPrefix+":Disable",
		blockerPrefix,
	))
}

func (g *generator) generateScheduler(name string, blocker AccessFilter) {
	blockerPrefix := g.IdentifierPrefix + ":" + name

	g.printSectionf("schedule scripts")

	g.writeLine("/system scheduler")

	for _, interval := range blocker.DisableIntervals {
		g.writeLine(fmt.Sprintf(
			`add name="%s" on-event="%s" interval=1d  start-time="%s"`,
			blockerPrefix+": Disable at "+interval.Start,
			blockerPrefix+":Disable",
			interval.Start,
		))
		g.writeLine(fmt.Sprintf(
			`add name="%s" on-event="%s" interval=1d  start-time="%s"`,
			blockerPrefix+": Enable  at "+interval.End,
			blockerPrefix+":Enable",
			interval.End,
		))
	}
}

func (g *generator) generatePrintInfo() {
	g.printSectionf("print configuration")
	g.writeLine("/system scheduler print")
	g.writeLine("/system script print")
	g.writeLine("/ip firewall filter print")
	g.writeLine("/ip firewall address-list print")
}

func (g *generator) printSectionf(s string, args ...interface{}) {
	s = fmt.Sprintf(s, args...)
	s = "# " + s + " #"

	b := make([]byte, len(s))
	for i := 0; i < len(s); i++ {
		b[i] = '#'
	}

	g.builder.WriteByte('\n')
	g.writeBytesLine(b)
	g.writeLine(s)
	g.writeBytesLine(b)
	g.builder.WriteByte('\n')
}

func (g *generator) writeBytesLine(b []byte) {
	g.builder.Write(b)
	g.builder.WriteByte('\n')
}

func (g *generator) writeLine(s string) {
	g.builder.WriteString(s)
	g.builder.WriteByte('\n')
}

func (g *generator) writeTable(rows [][]string) {
	cols := len(rows[0])
	colLengths := make([]int, cols)

	for _, row := range rows {
		for i := 0; i < cols; i++ {
			if len(row[i]) > colLengths[i] {
				colLengths[i] = len(row[i])
			}
		}
	}

	colPatterns := []string{}
	for _, colLength := range colLengths {
		colPatterns = append(colPatterns, fmt.Sprintf("%%-%ds", colLength))
	}
	rowPattern := strings.Join(colPatterns, " ")

	for _, row := range rows {
		asInterfaces := []interface{}{}
		for _, col := range row {
			asInterfaces = append(asInterfaces, col)
		}

		rowStr := fmt.Sprintf(rowPattern, asInterfaces...)
		rowStr = strings.TrimRight(rowStr, " ")

		g.writeLine(rowStr)
	}
}

func Generate(configuration Configuration) string {
	g := &generator{Configuration: configuration}

	return g.generate()
}
