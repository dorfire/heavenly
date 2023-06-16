package godepresolver

import (
	"github.com/dorfire/heavenly/pkg/goparse"
)

func (r *GoDepResolver) collectImports(pkgPath string, test, inclTransitive bool) (pkgImports, error) {
	if inclTransitive {
		return r.collectTransitiveImports(pkgPath, test, true)
	}
	return collectDirectImports(pkgPath, test)
}

// collectTransitiveImports returns the test/non-test imports of the given package,
// plus the non-test imports of its module-internal transitive deps.
func (r *GoDepResolver) collectTransitiveImports(pkgPath string, test, isFirstCall bool) (pkgImports, error) {
	// If imports for this pkgPath have already been collected in the past, return them
	if s, ok := r.pkgImportCache[pkgImportCacheKey(pkgPath, test)]; ok {
		return s, nil
	}

	imports, err := collectDirectImports(pkgPath, test)
	if err != nil {
		return nil, err
	}
	if !isFirstCall {
		for k, _ := range imports {
			imports[k] = true
		}
	}

	// for each module-internal package, recursively resolve its NON-TEST imports
	for dep, _ := range imports {
		if !r.isInternalImport(dep) {
			continue
		}
		r.log.DebugPrintf("Resolving transitive internal import %q from %q", dep, pkgPath)
		depImports, err := r.collectTransitiveImports(r.pkgGoPathToFSDir(dep), false, false)
		if err != nil {
			return nil, err
		}
		extend(imports, depImports)
	}

	r.pkgImportCache[pkgImportCacheKey(pkgPath, test)] = imports
	return imports, nil
}

func collectDirectImports(pkgPath string, test bool) (pkgImports, error) {
	imps, err := goparse.PackageImports(pkgPath, test)
	if err != nil {
		return nil, err
	}
	res := pkgImports{}
	for i := range imps.Iter() {
		res[i] = false
	}
	return res, nil
}
