package filtertest

import (
	"os"
	"path/filepath"
)

func fileFilters2() {
	importsTest(os.PathSeparator, "path/filepath") // want `YES`
	importsTest(filepath.Separator, "path/filepath")

	importsTest(os.PathListSeparator, "path/filepath") // want `YES`
	importsTest(filepath.ListSeparator, "path/filepath")
}
