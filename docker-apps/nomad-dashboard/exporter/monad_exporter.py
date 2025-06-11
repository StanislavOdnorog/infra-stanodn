#!/usr/bin/env python3

import json
import time
from prometheus_client import start_http_server, Gauge, Counter, Histogram, Summary
import requests
from typing import Dict, Any, Optional, List
import os
from dotenv import load_dotenv
import logging
import urllib3
from collections import deque
from datetime import datetime, timedelta
from urllib.parse import urlparse, parse_qs, urlencode

# Disable SSL warnings
urllib3.disable_warnings(urllib3.exceptions.InsecureRequestWarning)

# Configure logging
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(levelname)s - %(message)s'
)
logger = logging.getLogger(__name__)

# Load environment variables
load_dotenv()

# Prometheus metrics (module-level, not per-class)
network_version = Gauge('monad_network_version', 'Network version')
chain_id = Gauge('monad_chain_id', 'Chain ID')
is_syncing = Gauge('monad_is_syncing', 'Node syncing status')
block_number = Gauge('monad_block_number', 'Current block number')
block_gas_limit = Gauge('monad_block_gas_limit', 'Block gas limit')
block_gas_used = Gauge('monad_block_gas_used', 'Block gas used')
block_base_fee = Gauge('monad_block_base_fee', 'Block base fee per gas')
block_timestamp = Gauge('monad_block_timestamp', 'Block timestamp')
block_transaction_count = Gauge('monad_block_transaction_count', 'Number of transactions in block')
gas_price = Gauge('monad_gas_price', 'Current gas price in wei')
fee_history_base_fee = Gauge('monad_fee_history_base_fee', 'Base fee per gas from fee history')
fee_history_gas_used_ratio = Gauge('monad_fee_history_gas_used_ratio', 'Gas used ratio from fee history')
client_version = Gauge('monad_client_version', 'Node client version (as int hash)')
rpc_errors = Counter('monad_rpc_errors_total', 'Total number of RPC errors', ['method'])

def get_rpc_url() -> str:
    """Get RPC URL from environment variables and handle API key if present."""
    base_url = os.getenv("MONAD_RPC_URL")
    api_key = os.getenv("MONAD_RPC_KEY")
    
    if not api_key:
        return base_url
        
    # Parse the URL and add the API key
    parsed_url = urlparse(base_url)
    query_params = parse_qs(parsed_url.query)
    query_params['dkey'] = [api_key]
    
    # Reconstruct the URL with the API key
    new_query = urlencode(query_params, doseq=True)
    return f"{parsed_url.scheme}://{parsed_url.netloc}{parsed_url.path}?{new_query}"

class MonadExporter:
    def __init__(self, rpc_url, poll_interval=5):
        self.rpc_url = rpc_url
        self.poll_interval = poll_interval

    def make_rpc_call(self, method, params=None):
        if params is None:
            params = []
        payload = {
            "jsonrpc": "2.0",
            "id": 1,
            "method": method,
            "params": params
        }
        try:
            response = requests.post(self.rpc_url, json=payload)
            response.raise_for_status()
            data = response.json()
            if 'error' in data:
                logger.error(f"RPC error: {data['error']}")
                rpc_errors.labels(method=method).inc()
                return None
            return data.get('result')
        except Exception as e:
            logger.error(f"Error making RPC call: {str(e)}")
            if hasattr(e, 'response') and e.response is not None:
                logger.error(f"Response text: {e.response.text}")
            rpc_errors.labels(method=method).inc()
            return None

    def update_network_metrics(self):
        version = self.make_rpc_call('net_version')
        if version is not None:
            try:
                network_version.set(float(version))
            except Exception:
                pass
        chain_id_val = self.make_rpc_call('eth_chainId')
        if chain_id_val is not None:
            try:
                chain_id.set(float(int(chain_id_val, 16)))
            except Exception:
                pass
        syncing = self.make_rpc_call('eth_syncing')
        if syncing is not None:
            is_syncing.set(1 if syncing else 0)
        client_ver = self.make_rpc_call('web3_clientVersion')
        if client_ver is not None:
            # Hash the string to a float for Prometheus (not ideal, but allows tracking changes)
            client_version.set(float(abs(hash(client_ver)) % 1e12))

    def update_block_metrics(self):
        block_num_hex = self.make_rpc_call('eth_blockNumber')
        if block_num_hex is not None:
            try:
                block_number.set(float(int(block_num_hex, 16)))
            except Exception:
                pass
        block = self.make_rpc_call('eth_getBlockByNumber', ['latest', False])
        if block is not None:
            try:
                block_number.set(float(int(block['number'], 16)))
                block_gas_limit.set(float(int(block['gasLimit'], 16)))
                block_gas_used.set(float(int(block['gasUsed'], 16)))
                if 'baseFeePerGas' in block:
                    block_base_fee.set(float(int(block['baseFeePerGas'], 16)))
                block_timestamp.set(float(int(block['timestamp'], 16)))
                block_transaction_count.set(len(block['transactions']))
            except Exception:
                pass
        tx_count_hex = self.make_rpc_call('eth_getBlockTransactionCountByNumber', ['latest'])
        if tx_count_hex is not None:
            try:
                block_transaction_count.set(float(int(tx_count_hex, 16)))
            except Exception:
                pass

    def update_gas_metrics(self):
        gas_price_val = self.make_rpc_call('eth_gasPrice')
        if gas_price_val is not None:
            try:
                gas_price.set(float(int(gas_price_val, 16)))
            except Exception:
                pass

    def update_fee_history_metrics(self):
        fee_history = self.make_rpc_call('eth_feeHistory', [4, 'latest', [25, 75]])
        if fee_history is not None:
            try:
                if fee_history['baseFeePerGas']:
                    fee_history_base_fee.set(float(int(fee_history['baseFeePerGas'][-1], 16)))
                if fee_history['gasUsedRatio']:
                    fee_history_gas_used_ratio.set(float(fee_history['gasUsedRatio'][-1]))
            except Exception:
                pass

    def update_metrics(self):
        try:
            self.update_network_metrics()
            self.update_block_metrics()
            self.update_gas_metrics()
            self.update_fee_history_metrics()
        except Exception as e:
            logger.error(f"Error updating metrics: {str(e)}")

    def run(self):
        logger.info(f"Starting Monad exporter on port 8000")
        start_http_server(8000)
        while True:
            self.update_metrics()
            time.sleep(self.poll_interval)

def main():
    # Load environment variables from .env file
    load_dotenv(os.path.join(os.path.dirname(__file__), '.env'))
    
    # Create exporter instance
    rpc_url = get_rpc_url()
    poll_interval = int(os.getenv('POLL_INTERVAL', '5'))
    exporter = MonadExporter(rpc_url, poll_interval)
    
    # Start Prometheus metrics server
    exporter.run()

if __name__ == "__main__":
    main() 