#! /bin/bash

VERACITY_INSTALL=${VERACITY_INSTALL:-../../veracity}
DATATRAILS_URL=${DATATRAILS_URL:-https://app.datatrails.ai}
PUBLIC_ASSET_ID=${PUBLIC_ASSET_ID:-publicassets/87dd2e5a-42b4-49a5-8693-97f40a5af7f8}
PUBLIC_EVENT_ID=${PUBLIC_EVENT_ID:-publicassets/87dd2e5a-42b4-49a5-8693-97f40a5af7f8/events/a022f458-8e55-4d63-a200-4172a42fc2aa}

PROD_PUBLIC_LOGID=6ea5cd00-c711-3649-6914-7b125928bbb4
PROD_PUBLIC_TENANT_ID=${PROD_PUBLIC_TENANT_ID:-tenant/$PROD_PUBLIC_LOGID}

PROD_LOG_URL=${PROD_LOG_URL:-${DATATRAILS_URL}/verifiabledata/merklelogs/v1/mmrs/${PROD_PUBLIC_TENANT_ID}/0/massifs/0000000000000000.log}
TEST_TMPDIR=${TEST_TMPDIR:-${SHUNIT_TMPDIR}}
EMPTY_DIR=$TEST_TMPDIR/empty
PROD_DIR=$TEST_TMPDIR/prod
DUP_DIR=$TEST_TMPDIR/duplicate-massifs
PROD_LOCAL_BLOB_FILE="$PROD_DIR/mmr.log"
INVALID_BLOB_FILE="$TEST_TMPDIR/invalid.log"

oneTimeSetUp() {
  mkdir -p $EMPTY_DIR
  mkdir -p $PROD_DIR
  mkdir -p $DUP_DIR
  curl -s -H "x-ms-blob-type: BlockBlob" -H "x-ms-version: 2019-12-12" $PROD_LOG_URL -o $PROD_LOCAL_BLOB_FILE
  touch $INVALID_BLOB_FILE

  # Duplicate the prod and soak massif files in a single directory. The
  # replication should refuse to work with a directory that has multiple
  # massif files for the same massif index.
  cp $PROD_LOCAL_BLOB_FILE $DUP_DIR/prod-mmr.log

  assertTrue "prod MMR blob file should be present" "[ -r $PROD_LOCAL_BLOB_FILE ]"
  assertTrue "invalid MMR blob file should be present" "[ -r $INVALID_BLOB_FILE ]"
}

assertStringMatch() {
  local message="$1"
  local expected="$2"
  local actual="$3"

  # Normalize by converting all spaces to a single space, removing leading/trailing spaces and punctuation.
  expected=$(echo "$expected" | sed -e 's/[[:space:]]\+/ /g' -e 's/^[[:space:]]*//;s/[[:space:]]*[[:punct:]]*$//')
  actual=$(echo "$actual" | sed -e 's/[[:space:]]\+/ /g' -e 's/^[[:space:]]*//;s/[[:space:]]*[[:punct:]]*$//')

  echo "Expected (hex):" && echo "$expected" | hexdump -C
  echo "Actual (hex):" && echo "$actual" | hexdump -C

  assertEquals "$message" "$expected" "$actual"
}
