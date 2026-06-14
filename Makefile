SHELL := /bin/bash

ANSIBLE_PLAYBOOK ?= ansible-playbook
ANSIBLE_INVENTORY ?= ansible-inventory

STATIC_INVENTORY := ansible/inventory/static.yml,ansible/inventory/belle.yml
VPN_INVENTORY := ansible/inventory/vpn.yml

STATIC_HOST ?= vpn-ge
PROXMOX_HOST ?= pve01
TARGET ?=
APP_DIR ?=

.DEFAULT_GOAL := help

.PHONY: help inventory-graph ping server-init docker-install k3s-install app-deploy vpn-install template-ubuntu syntax-static syntax-proxmox check guard-TARGET guard-APP_DIR

help: ## Show available targets
	@grep -E '^[a-zA-Z0-9_-]+:.*## ' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*## "}; {printf "%-18s %s\n", $$1, $$2}'

inventory-graph: ## Show the static inventory graph
	$(ANSIBLE_INVENTORY) -i $(STATIC_INVENTORY) --graph

ping: guard-TARGET ## Ping a host from static inventory. Example: make ping TARGET=vpn-ge
	$(ANSIBLE_PLAYBOOK) -i $(STATIC_INVENTORY) ansible/playbooks/general/custom/ping.yml --extra-vars "target=$(TARGET)"

server-init: guard-TARGET ## Prepare a new or existing server with base users, packages, firewall, and hardening
	$(ANSIBLE_PLAYBOOK) -i $(STATIC_INVENTORY) ansible/playbooks/general/configure/server/init.yml --extra-vars "target=$(TARGET)"

docker-install: guard-TARGET ## Install Docker and Docker Compose on a host
	$(ANSIBLE_PLAYBOOK) -i $(STATIC_INVENTORY) ansible/playbooks/general/install/docker.yml --extra-vars "target=$(TARGET)"

k3s-install: guard-TARGET ## Install or reconcile a single-node k3s server on a host
	$(ANSIBLE_PLAYBOOK) -i $(STATIC_INVENTORY) ansible/playbooks/general/install/k3s.yml --extra-vars "target=$(TARGET)"

app-deploy: guard-TARGET guard-APP_DIR ## Deploy a docker-apps/<APP_DIR> project to a host
	$(ANSIBLE_PLAYBOOK) -i $(STATIC_INVENTORY) ansible/playbooks/general/install/custom-app.yml --extra-vars "target=$(TARGET) app_dir=$(APP_DIR)"

vpn-install: ## Deploy the VPN stack using ansible/inventory/vpn.yml
	$(ANSIBLE_PLAYBOOK) -i $(VPN_INVENTORY) ansible/playbooks/general/install/vpn.yml

template-ubuntu: guard-TARGET ## Create or update the Ubuntu cloud template on a Proxmox host
	$(ANSIBLE_PLAYBOOK) -i $(STATIC_INVENTORY) ansible/playbooks/proxmox/template/ubuntu.yml --extra-vars "target=$(TARGET)"

syntax-static: ## Syntax-check the main static-inventory playbooks
	$(ANSIBLE_PLAYBOOK) -i $(STATIC_INVENTORY) ansible/playbooks/general/custom/ping.yml --syntax-check --extra-vars "target=$(STATIC_HOST)"
	$(ANSIBLE_PLAYBOOK) -i $(STATIC_INVENTORY) ansible/playbooks/general/configure/server/init.yml --syntax-check --extra-vars "target=$(STATIC_HOST)"
	$(ANSIBLE_PLAYBOOK) -i $(STATIC_INVENTORY) ansible/playbooks/general/install/docker.yml --syntax-check --extra-vars "target=$(STATIC_HOST)"
	$(ANSIBLE_PLAYBOOK) -i $(STATIC_INVENTORY) ansible/playbooks/general/install/k3s.yml --syntax-check --extra-vars "target=$(STATIC_HOST)"
	$(ANSIBLE_PLAYBOOK) -i $(STATIC_INVENTORY) ansible/playbooks/general/install/custom-app.yml --syntax-check --extra-vars "target=$(STATIC_HOST) app_dir=example-app"

syntax-proxmox: ## Syntax-check the Proxmox template playbook against a static inventory host
	$(ANSIBLE_PLAYBOOK) -i $(STATIC_INVENTORY) ansible/playbooks/proxmox/template/ubuntu.yml --syntax-check --extra-vars "target=$(PROXMOX_HOST)"

check: syntax-static syntax-proxmox ## Run all syntax checks

guard-TARGET:
	@test -n "$(TARGET)" || (echo "TARGET is required. Example: make ping TARGET=vpn-ge" && exit 1)

guard-APP_DIR:
	@test -n "$(APP_DIR)" || (echo "APP_DIR is required. Example: make app-deploy TARGET=stan-gw01 APP_DIR=reverse-proxy" && exit 1)
