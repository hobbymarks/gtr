package cli

import "strings"

// parseLangPairToken parses a single argv token as SRC:TL or :TL, with optional
// multi-target TL1+TL2+…. It is conservative: if any segment fails validation,
// it returns ok=false so the token is treated as ordinary text.
func parseLangPairToken(s string) (source string, targets []string, ok bool) {
	s = strings.TrimSpace(s)
	i := strings.Index(s, ":")
	if i < 0 {
		return "", nil, false
	}
	left := strings.TrimSpace(s[:i])
	right := strings.TrimSpace(s[i+1:])
	if right == "" {
		return "", nil, false
	}
	for _, p := range strings.Split(right, "+") {
		p = strings.TrimSpace(p)
		if !looksLikeLangToken(p) {
			return "", nil, false
		}
		targets = append(targets, p)
	}
	if len(targets) == 0 {
		return "", nil, false
	}
	if left == "" {
		source = "auto"
	} else {
		if !looksLikeLangToken(left) && left != "auto" {
			return "", nil, false
		}
		source = left
	}
	return source, targets, true
}

func looksLikeLangToken(s string) bool {
	if s == "" {
		return false
	}
	if len(s) > 32 {
		return false
	}
	for _, r := range s {
		switch {
		case r >= 'a' && r <= 'z':
		case r >= 'A' && r <= 'Z':
		case r >= '0' && r <= '9':
		case r == '-':
		default:
			return false
		}
	}
	return true
}

// stripLeadingLangSpec removes args[0] when it is a language specification and
// both source and target were left at CLI defaults (neither flag was set by
// the user). Otherwise args are returned unchanged.
func stripLeadingLangSpec(args []string, sourceChanged, targetChanged bool) (rest []string, source string, targets []string, stripped bool) {
	if len(args) == 0 || sourceChanged || targetChanged {
		return args, "", nil, false
	}
	src, tgts, ok := parseLangPairToken(args[0])
	if !ok || len(tgts) == 0 {
		return args, "", nil, false
	}
	return args[1:], src, tgts, true
}
