package rbacconnect

func SplitProc(proc string) (svc string, pkg string) {
	s := proc
	if len(s) > 0 && s[0] == '/' {
		s = s[1:]
	}
	// s: "pkg.Service/Method"
	for i := 0; i < len(s); i++ {
		if s[i] == '/' {
			svc = s[:i] // "pkg.Service"
			break
		}
	}
	for i := 0; i < len(svc); i++ {
		if svc[i] == '.' {
			pkg = svc[:i] // "pkg"
			break
		}
	}
	return
}
