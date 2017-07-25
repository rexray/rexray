all: build

include ../../../test-framework-pkg.mk

################################################################################
##                                  SCHEMA                                    ##
################################################################################
SCHEMA_JSON := ../../../libstorage.json
SCHEMA_SRC := schema_generated.go

$(SCHEMA_SRC): $(SCHEMA_JSON)
	@echo generating $@
	@printf "package schema\n\nconst (\n" >$@; \
		printf "\t// JSONSchema is the libStorage API JSON schema\n" >>$@; \
		printf "\tJSONSchema = \`" >>$@; \
		sed -e 's/^//' $< >>$@; \
		printf "\`\n)\n" >>$@;

build: $(SCHEMA_SRC)

clean-schema:
	rm -f $(SCHEMA_SRC)

clean: clean-schema

.PHONY: clean-schema
