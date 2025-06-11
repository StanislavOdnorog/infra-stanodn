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

	// Current block metrics (latest block only)
	currentBlockGasUsed = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "monad_current_block_gas_used",
		Help: "Gas used in the current block",
	})

	currentBlockGasLimit = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "monad_current_block_gas_limit",
		Help: "Gas limit of the current block",
	})

	currentBlockGasUtilization = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "monad_current_block_gas_utilization_percent",
		Help: "Gas utilization percentage of the current block",
	})

	currentBlockSize = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "monad_current_block_size_bytes",
		Help: "Size of the current block in bytes",
	})

	currentBlockTransactionCount = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "monad_current_block_transaction_count",
		Help: "Number of transactions in the current block",
	})

	// Rolling averages (last 10 blocks)
	avgBlockTime = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "monad_avg_block_time_seconds",
		Help: "Average block time in seconds over last 10 blocks",
	})

	avgTransactionsPerBlock = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "monad_avg_transactions_per_block",
		Help: "Average transactions per block over last 10 blocks",
	})

	avgGasUtilization = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "monad_avg_gas_utilization_percent",
		Help: "Average gas utilization percentage over last 10 blocks",
	})

	avgGasPriceWei = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "monad_avg_gas_price_wei",
		Help: "Average gas price in wei over last 10 blocks",
	})

	// Transaction type metrics (aggregated)
	transactionTypes = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "monad_transaction_types_last_10_blocks",
		Help: "Count of different transaction types in last 10 blocks",
	}, []string{"type"})

	totalContractCalls = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "monad_contract_calls_last_10_blocks",
		Help: "Total number of contract calls in last 10 blocks",
	})

	totalSimpleTransfers = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "monad_simple_transfers_last_10_blocks",
		Help: "Total number of simple transfers in last 10 blocks",
	})

	avgUniqueAddressesPerBlock = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "monad_avg_unique_addresses_per_block",
		Help: "Average number of unique addresses per block over last 10 blocks",
	})

	totalValueTransferredWei = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "monad_total_value_transferred_wei_last_10_blocks",
		Help: "Total value transferred in last 10 blocks (wei)",
	})

	transactionsPerSecond = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "monad_transactions_per_second",
		Help: "Average transactions per second",
	})

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

// Create custom registry to avoid Go and process metrics
var customRegistry = prometheus.NewRegistry()

func init() {
	// Register only Monad-specific metrics with custom registry
	customRegistry.MustRegister(currentBlockNumber)
	customRegistry.MustRegister(gasPrice)
	customRegistry.MustRegister(gasPriceGwei)
	customRegistry.MustRegister(maxPriorityFeePerGas)
	customRegistry.MustRegister(currentBlockGasUsed)
	customRegistry.MustRegister(currentBlockGasLimit)
	customRegistry.MustRegister(currentBlockGasUtilization)
	customRegistry.MustRegister(currentBlockSize)
	customRegistry.MustRegister(currentBlockTransactionCount)
	customRegistry.MustRegister(avgBlockTime)
	customRegistry.MustRegister(avgTransactionsPerBlock)
	customRegistry.MustRegister(avgGasUtilization)
	customRegistry.MustRegister(avgGasPriceWei)
	customRegistry.MustRegister(transactionTypes)
	customRegistry.MustRegister(totalContractCalls)
	customRegistry.MustRegister(totalSimpleTransfers)
	customRegistry.MustRegister(avgUniqueAddressesPerBlock)
	customRegistry.MustRegister(totalValueTransferredWei)
	customRegistry.MustRegister(transactionsPerSecond)
	customRegistry.MustRegister(rpcResponseTime)
	customRegistry.MustRegister(rpcErrors)
	customRegistry.MustRegister(lastBlockTimestamp)
	customRegistry.MustRegister(metricsCollectionTime)
	customRegistry.MustRegister(chainInfo)
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
	const blocksToAnalyze = 5
	var totalTxs int
	var totalGasUsed, totalGasLimit uint64
	var totalContractCallsCount, totalSimpleTransfersCount int
	var totalUniqueAddresses int
	var totalValue, totalGasPrice uint64
	var totalGasPriceCount int
	txTypeCountTotal := make(map[string]int)

	for i := 0; i < blocksToAnalyze; i++ {
		blockNum := currentBlock - uint64(i)
		blockHex := fmt.Sprintf("0x%x", blockNum)
		
		// Add delay to avoid rate limiting
		if i > 0 {
			time.Sleep(100 * time.Millisecond)
		}
		
		block, err := c.GetBlock(blockHex, true)
		if err != nil {
			log.Printf("Failed to get block %d: %v", blockNum, err)
			// If rate limited, try to continue with fewer blocks
			if strings.Contains(err.Error(), "Too many request") {
				time.Sleep(500 * time.Millisecond)
			}
			continue
		}
		
		// Block metrics
		gasUsed := hexToUint64(block.GasUsed)
		gasLimit := hexToUint64(block.GasLimit)
		timestamp := hexToUint64(block.Timestamp)
		size := hexToUint64(block.Size)
		txCount := len(block.Transactions)

		// For current block (i == 0), set current metrics
		if i == 0 {
			currentBlockGasUsed.Set(float64(gasUsed))
			currentBlockGasLimit.Set(float64(gasLimit))
			currentBlockSize.Set(float64(size))
			currentBlockTransactionCount.Set(float64(txCount))

			if gasLimit > 0 {
				utilization := float64(gasUsed) / float64(gasLimit) * 100
				currentBlockGasUtilization.Set(utilization)
			}

			lastBlockTimestamp.Set(float64(timestamp))
		}

		// Accumulate for averages
		totalGasUsed += gasUsed
		totalGasLimit += gasLimit
		totalTxs += txCount

		// Store timestamp for block time calculation
		if i == 0 {
			// This is the current block, store its timestamp for comparison
			// We'll calculate block times outside this loop
		}

		// Transaction analysis
		if txCount > 0 {
			var contractCallCount, simpleTransferCount int
			uniqueAddrs := make(map[string]bool)

			for _, tx := range block.Transactions {
				// Gas price analysis
				txGasPrice := hexToUint64(tx.GasPrice)
				if txGasPrice == 0 {
					txGasPrice = hexToUint64(tx.MaxFeePerGas)
				}

				totalGasPrice += txGasPrice
				totalGasPriceCount++
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
				txTypeCountTotal[txType]++

				// Contract call vs simple transfer
				if tx.Input != "0x" && len(tx.Input) > 2 {
					contractCallCount++
				} else {
					simpleTransferCount++
				}
			}

			totalContractCallsCount += contractCallCount
			totalSimpleTransfersCount += simpleTransferCount
			totalUniqueAddresses += len(uniqueAddrs)
		}
	}

	// Set aggregated metrics
	if blocksToAnalyze > 0 {
		avgTransactionsPerBlock.Set(float64(totalTxs) / float64(blocksToAnalyze))
		
		if totalGasLimit > 0 {
			avgGasUtilization.Set(float64(totalGasUsed) / float64(totalGasLimit) * 100)
		}
		
		if totalGasPriceCount > 0 {
			avgGasPriceWei.Set(float64(totalGasPrice) / float64(totalGasPriceCount))
		}

		avgUniqueAddressesPerBlock.Set(float64(totalUniqueAddresses) / float64(blocksToAnalyze))
	}

	totalContractCalls.Set(float64(totalContractCallsCount))
	totalSimpleTransfers.Set(float64(totalSimpleTransfersCount))
	totalValueTransferredWei.Set(float64(totalValue))

	// Set transaction type metrics
	for txType, count := range txTypeCountTotal {
		transactionTypes.WithLabelValues(txType).Set(float64(count))
	}

	// Calculate average block time using a simpler approach
	// Get the first and last block timestamps to calculate average block time
	if blocksToAnalyze > 1 {
		firstBlockNum := currentBlock - uint64(blocksToAnalyze-1)
		firstBlockHex := fmt.Sprintf("0x%x", firstBlockNum)
		
		// Get first block with minimal data (no transactions)
		firstResult, err := c.makeRequest("eth_getBlockByNumber", []interface{}{firstBlockHex, false})
		if err == nil {
			firstBlockData, err := json.Marshal(firstResult)
			if err == nil {
				var firstBlockSimple struct {
					Timestamp string `json:"timestamp"`
				}
				if err := json.Unmarshal(firstBlockData, &firstBlockSimple); err == nil {
					firstTimestamp := hexToUint64(firstBlockSimple.Timestamp)
					
					// Get current block timestamp
					currentBlockHex := fmt.Sprintf("0x%x", currentBlock)
					currentResult, err := c.makeRequest("eth_getBlockByNumber", []interface{}{currentBlockHex, false})
					if err == nil {
						currentBlockData, err := json.Marshal(currentResult)
						if err == nil {
							var currentBlockSimple struct {
								Timestamp string `json:"timestamp"`
							}
							if err := json.Unmarshal(currentBlockData, &currentBlockSimple); err == nil {
								currentTimestamp := hexToUint64(currentBlockSimple.Timestamp)
								
								if currentTimestamp > firstTimestamp {
									totalTime := float64(currentTimestamp - firstTimestamp)
									avgBlockTimeValue := totalTime / float64(blocksToAnalyze-1)
									avgBlockTime.Set(avgBlockTimeValue)
									
									log.Printf("Block time calculation: Current block %d (ts: %d), First block %d (ts: %d), Total time: %f, Avg: %f", 
										currentBlock, currentTimestamp, firstBlockNum, firstTimestamp, totalTime, avgBlockTimeValue)
									
									// Calculate TPS
									if avgBlockTimeValue > 0 {
										avgTxsPerBlock := float64(totalTxs) / float64(blocksToAnalyze)
										tps := avgTxsPerBlock / avgBlockTimeValue
										transactionsPerSecond.Set(tps)
										log.Printf("TPS calculation: Total txs %d, Avg txs per block: %f, TPS: %f", totalTxs, avgTxsPerBlock, tps)
									}
								}
							}
						}
					}
				}
			}
		}
	}

	return nil
}

func main() {
	client := NewMonadClient("https://monad-testnet.drpc.org")

	// Set up HTTP server with custom registry (no Go/process metrics)
	http.Handle("/metrics", promhttp.HandlerFor(customRegistry, promhttp.HandlerOpts{}))
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
        <h1>Monad Testnet Prometheus Exporter</h1>
        
        <div class="info">
            <strong>Purpose:</strong> This service exports comprehensive Monad blockchain metrics in Prometheus format for monitoring and alerting.
        </div>

        <h2>Available Endpoints</h2>
        <ul>
            <li><a href="/metrics"><strong>/metrics</strong></a> - Prometheus metrics endpoint</li>
            <li><a href="/health"><strong>/health</strong></a> - Health check endpoint</li>
        </ul>

        <div class="metrics-list">
            <h3>Exported Metrics Categories</h3>
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
            <strong>Update Frequency:</strong> Metrics are collected every 30 seconds automatically.<br>
            <strong>Network:</strong> Monad Testnet (Chain ID: 10143)<br>
            <strong>RPC Endpoint:</strong> https://monad-testnet.drpc.org
        </div>

        <h3>Prometheus Configuration Example</h3>
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
			time.Sleep(60 * time.Second) // Collect metrics every 60 seconds
		}
	}()

	// Initial metrics collection
	log.Println("Starting Monad Testnet Prometheus Exporter...")
	if err := client.CollectMetrics(ctx); err != nil {
		log.Printf("Initial metrics collection failed: %v", err)
	} else {
		log.Println("Initial metrics collected successfully")
	}

	log.Println("Prometheus metrics available at: http://localhost:8080/metrics")
	log.Println("Health check available at: http://localhost:8080/health")
	log.Println("Dashboard available at: http://localhost:8080/")
	log.Println("Metrics update every 60 seconds")

	log.Fatal(http.ListenAndServe(":8080", nil))
} 