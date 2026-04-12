package llm

type Response struct {
	Answer    string
	Provider  string
	Model     string
	LatencyMS int
	TokensIn  *int
	TokensOut *int
	CostUSD   *float64
}
