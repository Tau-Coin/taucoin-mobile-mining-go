package ipfs

import (
	"os"
        "os/user"
        "path/filepath"
        "runtime"
)

// DefaultRepoPath is the default data directory to use for the ipfs persistence requirements.
func DefaultRepoPath() string {
	if repoPath := os.Getenv("TAU_IPFS_PATH"); repoPath != "" {
		return repoPath
	}

        // Try to place the data folder in the user's home dir
        home := homeDir()
        if home != "" {
                switch runtime.GOOS {
                case "darwin":
                        return filepath.Join(home, "Library", "Tau-ipfs")
                case "windows":
                        // We used to put everything in %HOME%\AppData\Roaming, but this caused
                        // problems with non-typical setups. If this fallback location exists and
                        // is non-empty, use it, otherwise DTRT and check %LOCALAPPDATA%.
                        fallback := filepath.Join(home, "AppData", "Roaming", "Tau-ipfs")
                        appdata := windowsAppData()
                        if appdata == "" || isNonEmptyDir(fallback) {
                                return fallback
                        }
                        return filepath.Join(appdata, "Tau-ipfs")
                default:
                        return filepath.Join(home, ".tau-ipfs")
                }
        }
        // As we cannot guess a stable location, return empty and handle later
        return ""
}

func homeDir() string {
        if home := os.Getenv("HOME"); home != "" {
                return home
        }
        if usr, err := user.Current(); err == nil {
                return usr.HomeDir
        }
        return ""
}

func windowsAppData() string {
        v := os.Getenv("LOCALAPPDATA")
        if v == "" {
                // Windows XP and below don't have LocalAppData. Crash here because
                // we don't support Windows XP and undefining the variable will cause
                // other issues.
                panic("environment variable LocalAppData is undefined")
        }
        return v
}

func isNonEmptyDir(dir string) bool {
        f, err := os.Open(dir)
        if err != nil {
                return false
        }
        names, _ := f.Readdir(1)
        f.Close()
        return len(names) > 0
}
