package cmd

// FIXME: Why do we have multiple "utils"?

func logErrorAndExit(err error) {
	if err != nil {
		panic(err)
		//fmt.Println(err)
		//os.Exit(1)
	}
}
