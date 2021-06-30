package gasserviceclient

type GasServiceClient interface {
	GetAccountAndSequence() (string, uint64, error)
	SubmitTxe(txeOfiRfiSignedB64 string) (string, uint64, error)
}
