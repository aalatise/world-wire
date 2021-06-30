package stellarconfig


type StellarConfigFinder interface {

    GetStellarConfigForDomain(domain string) (StellarConfig, error)

}


type StellarConfig struct {

    FederationServiceURL string `toml:"FEDERATION_SERVER"`
    ComplianceServiceURL string `toml:"AUTH_SERVER"`

}

