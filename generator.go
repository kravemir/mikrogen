package mikrogen

import (
	"fmt"
	"strings"
)

type generator struct {
	Configuration

	builder strings.Builder
}

func (g *generator) generate() string {
	g.generateOldCleanup()
	g.generateAddressList()
	g.generateFirewallFilter()
	g.generateToggleScripts()
	g.generateScheduler()
	g.generatePrintInfo()

	return g.builder.String()
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
			fmt.Sprintf(`[/ip firewall address-list find list~"%s*"]`, g.IdentifierPrefix),
		},
	})
}

func (g *generator) generateAddressList() {
	g.printSectionf("create address-list %s", g.IdentifierPrefix)

	g.writeLine("/ip firewall address-list")
	g.writeLine("")

	rows := [][]string{}
	for _, address := range g.DNSBlockedAddresses {
		rows = append(rows, []string{
			fmt.Sprintf("add address=%s", address),
			fmt.Sprintf("list=%s", g.IdentifierPrefix),
		})
	}
	g.writeTable(rows)
}

func (g *generator) generateFirewallFilter() {
	g.printSectionf("configure firewall filter rules")

	g.writeLine("/ip firewall filter")
	g.writeLine("")

	g.writeLine(fmt.Sprintf(
		`add comment="%s" action=reject chain=forward dst-address-list=blocked_web reject-with=icmp-network-unreachable`,
		g.IdentifierPrefix+":DNS",
	))
	g.writeLine("")

	for _, address := range g.TLSBlockedAddresses {
		g.writeLine(fmt.Sprintf(
			`add comment="%s" action=reject chain=forward protocol=tcp reject-with=icmp-network-unreachable tls-host="%s"`,
			g.IdentifierPrefix+":TLS",
			address,
		))
	}
}

func (g *generator) generateToggleScripts() {
	g.printSectionf("create scripts to enable / disable filters")

	g.writeLine("/system script")

	g.writeLine(fmt.Sprintf(
		`add name="%s" source="/foreach rule in=[/ip firewall filter find comment~\"%s*\"] do={ /ip firewall filter set \$rule disabled=no }"`,
		g.IdentifierPrefix+":Enable",
		g.IdentifierPrefix,
	))
	g.writeLine(fmt.Sprintf(
		`add name="%s" source="/foreach rule in=[/ip firewall filter find comment~\"%s*\"] do={ /ip firewall filter set \$rule disabled=yes }"`,
		g.IdentifierPrefix+":Disable",
		g.IdentifierPrefix,
	))
}

func (g *generator) generateScheduler() {
	g.printSectionf("schedule scripts")

	g.writeLine("/system scheduler")

	for _, interval := range g.DisableIntervals {
		g.writeLine(fmt.Sprintf(
			`add name="%s" on-event="%s" interval=1d  start-time="%s"`,
			g.IdentifierPrefix+": Disable at "+interval.Start,
			g.IdentifierPrefix+":Disable",
			interval.Start,
		))
		g.writeLine(fmt.Sprintf(
			`add name="%s" on-event="%s" interval=1d  start-time="%s"`,
			g.IdentifierPrefix+": Enable  at "+interval.Start,
			g.IdentifierPrefix+":Enable",
			interval.Start,
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
