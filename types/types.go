package types

type NetworkUrl string

const (
	MainNet    NetworkUrl = "https://bsc-dataseed.binance.org/"
	Testnet1s1 NetworkUrl = "https://data-seed-prebsc-1-s1.bnbchain.org:8545"
	Testnet2s1 NetworkUrl = "https://data-seed-prebsc-2-s1.bnbchain.org:8545"
	Testnet1s2 NetworkUrl = "https://data-seed-prebsc-1-s2.bnbchain.org:8545"
	Testnet2s2 NetworkUrl = "https://data-seed-prebsc-2-s2.bnbchain.org:8545"
	Testnet1s3 NetworkUrl = "https://data-seed-prebsc-1-s3.bnbchain.org:8545"
	Testnet2s3 NetworkUrl = "https://data-seed-prebsc-2-s3.bnbchain.org:8545"
)
