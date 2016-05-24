package version

var (
	SERVICE_DEFINITIONS = []string{
		//"arch/arm/bigchaindb.toml",
		//"arch/arm/btcd.toml",
		//"arch/arm/bitcoincore.toml",
		//"arch/arm/bitcoinclassic.toml",
		//"arch/arm/compilers.toml",
		//"arch/arm/geth.toml",
		"ipfs.toml",
		"keys.toml",
		//"arch/arm/logspout.toml",
		//"arch/arm/logrotate.toml",
		//"arch/arm/mindy.toml",
		//"arch/arm/openbazaar.toml",
		//"arch/arm/rethinkdb.toml",
		//"arch/arm/toadserver.toml",
		//"arch/arm/tinydns.toml",
		//"arch/arm/tor.toml",
		//"arch/arm/watchtower.toml",
		//"arch/arm/do_not_use.toml",
	}

	ACTION_DEFINITIONS = []string{
		"chain_info.toml",
		"dns_register.toml",
		"keys_list.toml",
	}

	CHAIN_DEFINITIONS = []string{
		"default.toml",
		"config.toml",
		"server_conf.toml",
	}
)
