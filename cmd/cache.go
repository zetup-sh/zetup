package cmd

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

// cacheCmd represents the cache command
var cacheCmd = &cobra.Command{
	Use:   "cache",
	Short: "simple caching system for zetup",
	Long: `all items are stored 1 level deep as strings
	usage: zetup cache set [key] [item]
	or: zetup cache get [key]

	This persists throughout all installations, so if you need something only in your installation, prefix with a unique identifier like the name of your zetup package.

	`,
	Run: func(cmd *cobra.Command, args []string) {
		cacheOutputAll()
	},
}

var getCacheCmd = &cobra.Command{
	Use:   "get",
	Args:  cobra.ExactArgs(1),
	Short: "get cache item",
	Long: `all items are stored 1 level deep as strings
	zetup cache get [key]
	outputs empty string if not set
	`,
	Run: func(cmd *cobra.Command, args []string) {
		cacheGet(args[0])
	},
}

var setCacheCmd = &cobra.Command{
	Use:   "set",
	Args:  cobra.ExactArgs(2),
	Short: "set cache item",
	Long: `all items are stored 1 level deep as strings
	zetup cache set [key] [item]
	`,
	Run: func(cmd *cobra.Command, args []string) {
		cacheSet(args[0], args[1])
	},
}

var clearCacheCmd = &cobra.Command{
	Use:   "clear",
	Args:  cobra.ExactArgs(0),
	Short: "completely destroys everything in cache",
	Long:  `deletes cache file`,
	Run: func(cmd *cobra.Command, args []string) {
		cacheClear()
	},
}

var cacheFile string

func init() {
	rootCmd.AddCommand(cacheCmd)
	cacheCmd.PersistentFlags().StringVarP(&cacheFile, "cache-file", "f", "", "file where to store cache")
	cacheCmd.AddCommand(getCacheCmd)
	cacheCmd.AddCommand(setCacheCmd)
	cacheCmd.AddCommand(clearCacheCmd)

	cobra.OnInitialize(func() {
		if cacheFile == "" {
			cacheFile = filepath.Join(zetupDir, ".cache")
		}
	})
}

type tCacheItems map[string]string

func cacheGet(key string) {
	cacheItems := readCacheFile()
	if val, ok := cacheItems[key]; ok {
		fmt.Printf(val)
	} else {
		fmt.Printf("")
	}
}

func cacheSet(key string, item string) {
	cacheItems := readCacheFile()
	cacheItems[key] = item
	marshaled, err := yaml.Marshal(cacheItems)
	check(err)

	cacheWithHeader := []byte("# generated file do not edit\n" + string(marshaled))
	err = ioutil.WriteFile(cacheFile, cacheWithHeader, 0644)
	check(err)
}

func cacheOutputAll() {
	cacheItems := readCacheFile()
	for key, item := range cacheItems {
		fmt.Println(key, item)
	}
}

func readCacheFile() tCacheItems {
	var cacheItems tCacheItems
	if exists(cacheFile) {
		dat, err := ioutil.ReadFile(cacheFile)
		check(err)
		yaml.Unmarshal(dat, &cacheItems)
		return cacheItems
	}
	return make(tCacheItems)
}

func cacheClear() {
	os.Remove(cacheFile)
}
