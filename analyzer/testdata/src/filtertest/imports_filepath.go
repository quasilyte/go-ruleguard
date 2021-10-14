package filtertest

import (
	"os"
	"path/filepath"
)

func fileFilters2() {
	importsTest(os.PathSeparator, "path/filepath") // want `true`
	importsTest(filepath.Separator, "path/filepath")

	importsTest(os.PathListSeparator, "path/filepath") // want `true`
	importsTest(filepath.ListSeparator, "path/filepath")
}
