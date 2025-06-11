package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type MonadClient struct {
	URL string
}

type JSONRPCRequest struct {
	JSONRPC string      `json:"jsonrpc"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params"`
	ID      int         `json:"id"`
}

type JSONRPCResponse struct {
	JSONRPC string      `json:"jsonrpc"`
	Result  interface{} `json:"result"`
	Error   interface{} `json:"error"`
	ID      int         `json:"id"`
}

type Block struct {
	Number           string        `json:"number"`
	Hash             string        `json:"hash"`
	ParentHash       string        `json:"parentHash"`
	Timestamp        string        `json:"timestamp"`
	GasLimit         string        `json:"gasLimit"`
	GasUsed          string        `json:"gasUsed"`
	BaseFeePerGas    string        `json:"baseFeePerGas"`
	Transactions     []Transaction `json:"transactions"`
	Size             string        `json:"size"`
	Difficulty       string        `json:"difficulty"`
	TotalDifficulty  string        `json:"totalDifficulty"`
	Miner            string        `json:"miner"`
	ExtraData        string        `json:"extraData"`
}

type Transaction struct {
	Hash                 string `json:"hash"`
	From                 string `json:"from"`
	To                   string `json:"to"`
	Value                string `json:"value"`
	Gas                  string `json:"gas"`
	GasPrice             string `json:"gasPrice"`
	MaxFeePerGas         string `json:"maxFeePerGas"`
	MaxPriorityFeePerGas string `json:"maxPriorityFeePerGas"`
	Input                string `json:"input"`
	Type                 string `json:"type"`
	Nonce                string `json:"nonce"`
	TransactionIndex     string `json:"transactionIndex"`
	BlockNumber          string `json:"blockNumber"`
	BlockHash            string `json:"blockHash"`
}

// Prometheus metrics
var (
	// Network metrics
	currentBlockNumber = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "monad_current_block_number",
		Help: "Current block number on the Monad network",
	})

	gasPrice = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "monad_gas_price_wei",
		Help: "Current gas price in wei",
	})

	gasPriceGwei = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "monad_gas_price_gwei",
		Help: "Current gas price in gwei",
	})

	maxPriorityFeePerGas = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "monad_max_priority_fee_per_gas_wei",
		Help: "Current max priority fee per gas in wei",
	})

	// Block metrics
	blockTime = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "monad_avg_block_time_seconds",
		Help: "Average block time in seconds over recent blocks",
	})

	transactionsPerSecond = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "monad_transactions_per_second",
		Help: "Average transactions per second",
	})

	blockGasUsed = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "monad_block_gas_used",
		Help: "Gas used in the latest block",
	}, []string{"block_number"})

	blockGasLimit = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "monad_block_gas_limit",
		Help: "Gas limit of the latest block",
	}, []string{"block_number"})

	blockGasUtilization = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "monad_block_gas_utilization_percent",
		Help: "Gas utilization percentage of the latest block",
	}, []string{"block_number"})

	blockSize = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "monad_block_size_bytes",
		Help: "Size of the latest block in bytes",
	}, []string{"block_number"})

	blockTransactionCount = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "monad_block_transaction_count",
		Help: "Number of transactions in the latest block",
	}, []string{"block_number"})

	// Transaction metrics
	transactionTypes = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "monad_transaction_types",
		Help: "Count of different transaction types in recent blocks",
	}, []string{"type", "block_number"})

	contractCalls = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "monad_contract_calls",
		Help: "Number of contract calls in recent blocks",
	}, []string{"block_number"})

	simpleTransfers = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "monad_simple_transfers",
		Help: "Number of simple transfers in recent blocks",
	}, []string{"block_number"})

	uniqueAddresses = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "monad_unique_addresses",
		Help: "Number of unique addresses in recent blocks",
	}, []string{"block_number"})

	totalValueTransferred = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "monad_total_value_transferred_wei",
		Help: "Total value transferred in recent blocks (wei)",
	}, []string{"block_number"})

	// Gas price statistics
	avgGasPrice = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "monad_avg_gas_price_per_block_wei",
		Help: "Average gas price per block in wei",
	}, []string{"block_number"})

	maxGasPrice = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "monad_max_gas_price_per_block_wei",
		Help: "Maximum gas price per block in wei",
	}, []string{"block_number"})

	minGasPrice = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "monad_min_gas_price_per_block_wei",
		Help: "Minimum gas price per block in wei",
	}, []string{"block_number"})

	// Network health metrics
	rpcResponseTime = prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    "monad_rpc_response_time_seconds",
		Help:    "Response time for RPC calls",
		Buckets: prometheus.DefBuckets,
	})

	rpcErrors = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "monad_rpc_errors_total",
		Help: "Total number of RPC errors by method",
	}, []string{"method"})

	lastBlockTimestamp = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "monad_last_block_timestamp",
		Help: "Timestamp of the last processed block",
	})

	metricsCollectionTime = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "monad_metrics_collection_time_seconds",
		Help: "Time taken to collect all metrics",
	})

	// Chain information
	chainInfo = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "monad_chain_info",
		Help: "Chain information (value always 1, labels contain info)",
	}, []string{"chain_id", "client_version", "network"})
)

func init() {
	// Register all metrics
	prometheus.MustRegister(currentBlockNumber)
	prometheus.MustRegister(gasPrice)
	prometheus.MustRegister(gasPriceGwei)
	prometheus.MustRegister(maxPriorityFeePerGas)
	prometheus.MustRegister(blockTime)
	prometheus.MustRegister(transactionsPerSecond)
	prometheus.MustRegister(blockGasUsed)
	prometheus.MustRegister(blockGasLimit)
	prometheus.MustRegister(blockGasUtilization)
	prometheus.MustRegister(blockSize)
	prometheus.MustRegister(blockTransactionCount)
	prometheus.MustRegister(transactionTypes)
	prometheus.MustRegister(contractCalls)
	prometheus.MustRegister(simpleTransfers)
	prometheus.MustRegister(uniqueAddresses)
	prometheus.MustRegister(totalValueTransferred)
	prometheus.MustRegister(avgGasPrice)
	prometheus.MustRegister(maxGasPrice)
	prometheus.MustRegister(minGasPrice)
	prometheus.MustRegister(rpcResponseTime)
	prometheus.MustRegister(rpcErrors)
	prometheus.MustRegister(lastBlockTimestamp)
	prometheus.MustRegister(metricsCollectionTime)
	prometheus.MustRegister(chainInfo)
}

func NewMonadClient(url string) *MonadClient {
	return &MonadClient{URL: url}
}

func (c *MonadClient) makeRequest(method string, params interface{}) (interface{}, error) {
	start := time.Now()
	defer func() {
		rpcResponseTime.Observe(time.Since(start).Seconds())
	}()

	request := JSONRPCRequest{
		JSONRPC: "2.0",
		Method:  method,
		Params:  params,
		ID:      1,
	}

	requestBody, err := json.Marshal(request)
	if err != nil {
		rpcErrors.WithLabelValues(method).Inc()
		return nil, err
	}

	resp, err := http.Post(c.URL, "application/json", strings.NewReader(string(requestBody)))
	if err != nil {
		rpcErrors.WithLabelValues(method).Inc()
		return nil, err
	}
	defer resp.Body.Close()

	var response JSONRPCResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		rpcErrors.WithLabelValues(method).Inc()
		return nil, err
	}

	if response.Error != nil {
		rpcErrors.WithLabelValues(method).Inc()
		return nil, fmt.Errorf("RPC error: %v", response.Error)
	}

	return response.Result, nil
}

func (c *MonadClient) GetChainID() (string, error) {
	result, err := c.makeRequest("eth_chainId", []interface{}{})
	if err != nil {
		return "", err
	}
	return result.(string), nil
}

func (c *MonadClient) GetBlockNumber() (uint64, error) {
	result, err := c.makeRequest("eth_blockNumber", []interface{}{})
	if err != nil {
		return 0, err
	}
	hexStr := result.(string)
	return strconv.ParseUint(hexStr[2:], 16, 64)
}

func (c *MonadClient) GetBlock(blockNumber string, fullTx bool) (*Block, error) {
	result, err := c.makeRequest("eth_getBlockByNumber", []interface{}{blockNumber, fullTx})
	if err != nil {
		return nil, err
	}

	blockData, err := json.Marshal(result)
	if err != nil {
		return nil, err
	}

	var block Block
	if err := json.Unmarshal(blockData, &block); err != nil {
		return nil, err
	}

	return &block, nil
}

func (c *MonadClient) GetGasPrice() (uint64, error) {
	result, err := c.makeRequest("eth_gasPrice", []interface{}{})
	if err != nil {
		return 0, err
	}
	hexStr := result.(string)
	return strconv.ParseUint(hexStr[2:], 16, 64)
}

func (c *MonadClient) GetMaxPriorityFeePerGas() (uint64, error) {
	result, err := c.makeRequest("eth_maxPriorityFeePerGas", []interface{}{})
	if err != nil {
		return 0, err
	}
	hexStr := result.(string)
	return strconv.ParseUint(hexStr[2:], 16, 64)
}

func (c *MonadClient) GetClientVersion() (string, error) {
	result, err := c.makeRequest("web3_clientVersion", []interface{}{})
	if err != nil {
		return "", err
	}
	return result.(string), nil
}

func hexToUint64(hexStr string) uint64 {
	if hexStr == "" || hexStr == "0x" {
		return 0
	}
	val, _ := strconv.ParseUint(hexStr[2:], 16, 64)
	return val
}

func (c *MonadClient) CollectMetrics(ctx context.Context) error {
	start := time.Now()
	defer func() {
		metricsCollectionTime.Set(time.Since(start).Seconds())
	}()

	// Get basic network info
	chainID, err := c.GetChainID()
	if err != nil {
		return fmt.Errorf("failed to get chain ID: %w", err)
	}

	clientVersion, err := c.GetClientVersion()
	if err != nil {
		return fmt.Errorf("failed to get client version: %w", err)
	}

	// Set chain info
	chainIDDecimal := hexToUint64(chainID)
	chainInfo.WithLabelValues(fmt.Sprintf("%d", chainIDDecimal), clientVersion, "monad-testnet").Set(1)

	// Get current block number
	currentBlock, err := c.GetBlockNumber()
	if err != nil {
		return fmt.Errorf("failed to get current block: %w", err)
	}
	currentBlockNumber.Set(float64(currentBlock))

	// Get gas price
	currentGasPrice, err := c.GetGasPrice()
	if err != nil {
		return fmt.Errorf("failed to get gas price: %w", err)
	}
	gasPrice.Set(float64(currentGasPrice))
	gasPriceGwei.Set(float64(currentGasPrice) / 1e9)

	// Get max priority fee (optional)
	if maxPriorityFee, err := c.GetMaxPriorityFeePerGas(); err == nil {
		maxPriorityFeePerGas.Set(float64(maxPriorityFee))
	}

	// Analyze recent blocks for detailed metrics
	const blocksToAnalyze = 10
	var blockTimes []float64
	var totalTxs int

	for i := 0; i < blocksToAnalyze; i++ {
		blockNum := currentBlock - uint64(i)
		blockHex := fmt.Sprintf("0x%x", blockNum)
		
		block, err := c.GetBlock(blockHex, true)
		if err != nil {
			log.Printf("Failed to get block %d: %v", blockNum, err)
			continue
		}

		blockNumStr := fmt.Sprintf("%d", blockNum)
		
		// Block metrics
		gasUsed := hexToUint64(block.GasUsed)
		gasLimit := hexToUint64(block.GasLimit)
		timestamp := hexToUint64(block.Timestamp)
		size := hexToUint64(block.Size)
		txCount := len(block.Transactions)

		blockGasUsed.WithLabelValues(blockNumStr).Set(float64(gasUsed))
		blockGasLimit.WithLabelValues(blockNumStr).Set(float64(gasLimit))
		blockSize.WithLabelValues(blockNumStr).Set(float64(size))
		blockTransactionCount.WithLabelValues(blockNumStr).Set(float64(txCount))

		if gasLimit > 0 {
			utilization := float64(gasUsed) / float64(gasLimit) * 100
			blockGasUtilization.WithLabelValues(blockNumStr).Set(utilization)
		}

		// Set last block timestamp for the most recent block
		if i == 0 {
			lastBlockTimestamp.Set(float64(timestamp))
		}

		// Calculate block time
		if i > 0 {
			prevBlockNum := currentBlock - uint64(i-1)
			prevBlockHex := fmt.Sprintf("0x%x", prevBlockNum)
			prevBlock, err := c.GetBlock(prevBlockHex, false)
			if err == nil {
				prevTimestamp := hexToUint64(prevBlock.Timestamp)
				if timestamp < prevTimestamp {
					blockTimeDiff := float64(prevTimestamp - timestamp)
					blockTimes = append(blockTimes, blockTimeDiff)
				}
			}
		}

		totalTxs += txCount

		// Transaction analysis
		if txCount > 0 {
			var totalGasPrice, totalValue uint64
			var maxGas, minGas uint64
			var contractCallCount, simpleTransferCount int
			uniqueAddrs := make(map[string]bool)
			txTypeCount := make(map[string]int)

			for j, tx := range block.Transactions {
				// Gas price analysis
				txGasPrice := hexToUint64(tx.GasPrice)
				if txGasPrice == 0 {
					txGasPrice = hexToUint64(tx.MaxFeePerGas)
				}

				if j == 0 || txGasPrice > maxGas {
					maxGas = txGasPrice
				}
				if j == 0 || txGasPrice < minGas {
					minGas = txGasPrice
				}

				totalGasPrice += txGasPrice
				totalValue += hexToUint64(tx.Value)

				// Address tracking
				uniqueAddrs[tx.From] = true
				if tx.To != "" {
					uniqueAddrs[tx.To] = true
				}

				// Transaction type analysis
				txType := "legacy"
				if tx.Type != "" {
					switch hexToUint64(tx.Type) {
					case 0:
						txType = "legacy"
					case 1:
						txType = "eip2930"
					case 2:
						txType = "eip1559"
					default:
						txType = "unknown"
					}
				}
				txTypeCount[txType]++

				// Contract call vs simple transfer
				if tx.Input != "0x" && len(tx.Input) > 2 {
					contractCallCount++
				} else {
					simpleTransferCount++
				}
			}

			// Set metrics
			avgGasPrice.WithLabelValues(blockNumStr).Set(float64(totalGasPrice / uint64(txCount)))
			maxGasPrice.WithLabelValues(blockNumStr).Set(float64(maxGas))
			minGasPrice.WithLabelValues(blockNumStr).Set(float64(minGas))
			contractCalls.WithLabelValues(blockNumStr).Set(float64(contractCallCount))
			simpleTransfers.WithLabelValues(blockNumStr).Set(float64(simpleTransferCount))
			uniqueAddresses.WithLabelValues(blockNumStr).Set(float64(len(uniqueAddrs)))
			totalValueTransferred.WithLabelValues(blockNumStr).Set(float64(totalValue))

			// Transaction type metrics
			for txType, count := range txTypeCount {
				transactionTypes.WithLabelValues(txType, blockNumStr).Set(float64(count))
			}
		}
	}

	// Calculate average block time and TPS
	if len(blockTimes) > 0 {
		var sum float64
		for _, bt := range blockTimes {
			sum += bt
		}
		avgBlockTime := sum / float64(len(blockTimes))
		blockTime.Set(avgBlockTime)

		// Calculate TPS
		if avgBlockTime > 0 {
			avgTxsPerBlock := float64(totalTxs) / float64(blocksToAnalyze)
			tps := avgTxsPerBlock / avgBlockTime
			transactionsPerSecond.Set(tps)
		}
	}

	return nil
}

func main() {
	client := NewMonadClient("https://monad-testnet.drpc.org")

	// Set up HTTP server
	http.Handle("/metrics", promhttp.Handler())
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(`
<!DOCTYPE html>
<html>
<head>
    <title>Monad Testnet Prometheus Exporter</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 40px; background: #f5f5f5; }
        .container { background: white; padding: 30px; border-radius: 8px; box-shadow: 0 2px 10px rgba(0,0,0,0.1); }
        h1 { color: #333; border-bottom: 2px solid #667eea; padding-bottom: 10px; }
        .info { background: #e3f2fd; padding: 15px; border-radius: 5px; margin: 20px 0; }
        a { color: #667eea; text-decoration: none; }
        a:hover { text-decoration: underline; }
        .metrics-list { background: #f8f9fa; padding: 15px; border-radius: 5px; margin: 20px 0; }
        .metrics-list h3 { margin-top: 0; color: #495057; }
        ul { margin: 10px 0; }
        li { margin: 5px 0; }
    </style>
</head>
<body>
    <div class="container">
        <h1>🚀 Monad Testnet Prometheus Exporter</h1>
        
        <div class="info">
            <strong>🎯 Purpose:</strong> This service exports comprehensive Monad blockchain metrics in Prometheus format for monitoring and alerting.
        </div>

        <h2>📊 Available Endpoints</h2>
        <ul>
            <li><a href="/metrics"><strong>/metrics</strong></a> - Prometheus metrics endpoint</li>
            <li><a href="/health"><strong>/health</strong></a> - Health check endpoint</li>
        </ul>

        <div class="metrics-list">
            <h3>📈 Exported Metrics Categories</h3>
            <ul>
                <li><strong>Network Metrics:</strong> Block height, gas prices, block times, TPS</li>
                <li><strong>Block Analysis:</strong> Gas utilization, transaction counts, block sizes</li>
                <li><strong>Transaction Types:</strong> Contract calls, transfers, EIP-1559 vs legacy</li>
                <li><strong>Gas Price Statistics:</strong> Min/max/average gas prices per block</li>
                <li><strong>Network Health:</strong> RPC response times, error rates</li>
                <li><strong>Address Activity:</strong> Unique addresses per block</li>
                <li><strong>Value Transfers:</strong> Total value moved per block</li>
            </ul>
        </div>

        <div class="info">
            <strong>🔄 Update Frequency:</strong> Metrics are collected every 30 seconds automatically.<br>
            <strong>🌐 Network:</strong> Monad Testnet (Chain ID: 10143)<br>
            <strong>📡 RPC Endpoint:</strong> https://monad-testnet.drpc.org
        </div>

        <h3>🐳 Prometheus Configuration Example</h3>
        <pre style="background: #f8f9fa; padding: 15px; border-radius: 5px; overflow-x: auto;">
scrape_configs:
  - job_name: 'monad-exporter'
    static_configs:
      - targets: ['localhost:8080']
    scrape_interval: 30s
    metrics_path: /metrics</pre>
    </div>
</body>
</html>
		`))
	})

	// Start metrics collection in background
	ctx := context.Background()
	go func() {
		for {
			if err := client.CollectMetrics(ctx); err != nil {
				log.Printf("Error collecting metrics: %v", err)
			}
			time.Sleep(30 * time.Second) // Collect metrics every 30 seconds
		}
	}()

	// Initial metrics collection
	log.Println("🚀 Starting Monad Testnet Prometheus Exporter...")
	if err := client.CollectMetrics(ctx); err != nil {
		log.Printf("⚠️  Initial metrics collection failed: %v", err)
	} else {
		log.Println("✅ Initial metrics collected successfully")
	}

	log.Println("📊 Prometheus metrics available at: http://localhost:8080/metrics")
	log.Println("🏥 Health check available at: http://localhost:8080/health")
	log.Println("🌐 Dashboard available at: http://localhost:8080/")
	log.Println("🔄 Metrics update every 30 seconds")

	log.Fatal(http.ListenAndServe(":8080", nil))
} 