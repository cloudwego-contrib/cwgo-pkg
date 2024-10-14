TOOLS_SHELL="./hack/tools.sh"

.PHONY: test
test:
	chmod +x ${TOOLS_SHELL}
	@${TOOLS_SHELL} test
	@echo "go test finished"



.PHONY: vet
vet:
	chmod +x ${TOOLS_SHELL}
	@${TOOLS_SHELL} vet
	@echo "vet check finished"