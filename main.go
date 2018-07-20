package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"syscall"

	"github.com/dbaumgarten/addfs/afs"
)

var usagestring = `Usage: addfs [flags] targetdir mountpoint

Mount a directory as read+append only. Inside the mountpoint new files and folders can be created, but existing files and folders can not be deleted, unlinked, renamed, moved or overwritten. 

ATTENTION: If a user has write-access to the real directory (not only the mountpoint) he can circumvent the write-protection by writing directly to the real directory rather then the mountpoint. Use correct directory-permissions to prevent this!!!

Possible flags:
`

var allowRootMutation = flag.Bool("allowRootMutation", false, "Allow the root-user to mutate files and folders")
var ignoreWarnings = flag.Bool("ignoreWarnings", false, "Ignore all errors about file-ownership. ONLY USE IF YOU KNOW WHAT YOU ARE DOING!")
var mutableFiles arrayFlags

func main() {
	flag.Var(&mutableFiles, "mutableFiles", "Allow mutation of files that match the given regex. Can be specified multiple times.")
	flag.Usage = func() {
		fmt.Println(usagestring)
		flag.PrintDefaults()
	}
	flag.Parse()
	if flag.NArg() != 2 {
		flag.Usage()
		os.Exit(1)
	}

	if !*ignoreWarnings {
		if os.Getegid() != 0 {
			fmt.Println("WARNING!: Not running. Only directories that are write-accessible for the user can be mounted. This defeats  addfs' write-protection. Aborting for safety-reasons.")
			os.Exit(1)
		}
		err := checkSourceDirectory(flag.Arg(0))
		if err != nil {
			fmt.Println("WARNING!: " + err.Error() + ". Aborting for safety-reasons.")
			os.Exit(1)
		}
	}

	afsOpts := afs.AddFSOpts{
		AllowRootMutation: *allowRootMutation,
		MutableFiles:      mutableFiles,
	}
	afs, err := afs.NewAddFS(flag.Arg(0), afsOpts)
	if err != nil {
		fmt.Println("Error: ", err)
		os.Exit(1)
	}
	fmt.Println("Starting FUSE-Filesystem!")
	err = afs.Run()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func checkSourceDirectory(path string) error {
	stat, err := os.Stat(path)
	if err != nil {
		if os.IsExist(err) {
			return errors.New("the source directory does not exists")
		}
		return err
	}
	sstat, ok := stat.Sys().(*syscall.Stat_t)
	if !ok {
		panic("Not a syscall.Stat_t")
	}
	if sstat.Uid != 0 {
		return errors.New("the source directory is not owned by root. A regular user could write to it directly and therefore defeat addfs' write-protection")
	}
	if 02&stat.Mode() != 0 {
		return errors.New("the source directory is world-writeable. A regular user could write to it directly and therefore defeat addfs' write-protection")
	}

	return nil
}
