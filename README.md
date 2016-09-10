# gconfig
Spring boot style configuration management for Go

Usage

    import "github.com/narup/gconfig"
    
    //load configuration
	if _, err := gconfig.Load(); err != nil {
		fmt.Printf("Error::%s\n", err.Error())
	}
	cfg = gconfig.Gcg

	//setup database
	host := cfg.GetString("maindb.host") + ":" + cfg.GetString("maindb.port")
	dbConfig := gmgo.DbConfig{Host: host, DBName: cfg.GetString("maindb.dbName"), UserName: "", Password: ""}
	if _, err := data.Setup(dbConfig); err != nil {
		log.Panicf("Error connecting to the database %s", err.Error())
	}