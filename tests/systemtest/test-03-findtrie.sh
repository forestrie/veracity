testFindTrieEntrySingleEvent() {
  # Verify the trie key for the known event is on the log at the correct position.
  PUBLIC_EVENT_PERMISSIONED_ID=${PUBLIC_EVENT_ID#"public"}
  output=$($VERACITY_INSTALL find-trie-entries --log-tenant $PROD_PUBLIC_TENANT_ID --app-id $PUBLIC_EVENT_PERMISSIONED_ID)
  assertEquals "verifying finding the trie entry of a known public prod event from the datatrails log should match mmr index 663" "matches: [663]" "$output"
}

testFindTrieEntrySingleEventWithLocalMassifCopy() {

  # Verify the trie key for the known event is on the log at the correct position for a local log.
  PUBLIC_EVENT_PERMISSIONED_ID=${PUBLIC_EVENT_ID#"public"}
  output=$($VERACITY_INSTALL --massif-file $PROD_LOCAL_BLOB_FILE find-trie-entries --log-tenant $PROD_PUBLIC_TENANT_ID --app-id $PUBLIC_EVENT_PERMISSIONED_ID)
  assertEquals "verifying finding the trie entry of a known public prod event from a local log should match mmr index 663" "matches: [663]" "$output"
}

testFindMMREntrySingleEvent() {
  # Verify the mmr entry for the known event is on the log at the correct position.
  echo "disabled, datatrails public events service has been shutdown"
  if false; then

    output=$(curl -sL $DATATRAILS_URL/archivist/v2/$PUBLIC_EVENT_ID | VERACITY_IKWID=true $VERACITY_INSTALL find-mmr-entries --log-tenant $PROD_PUBLIC_TENANT_ID)
    assertEquals "verifying finding the mmr entry of a known public prod event from the datatrails log should match mmr index 663" "matches: [663]" "$output"
  fi
}

testFindMMREntrySingleEventWithLocalMassifCopy() {
  echo "disabled, datatrails public events service has been shutdown"
  if false; then

    # Verify the mmr entry for the known event is on the log at the correct position.
    output=$(curl -sL $DATATRAILS_URL/archivist/v2/$PUBLIC_EVENT_ID | VERACITY_IKWID=true $VERACITY_INSTALL --data-local $PROD_LOCAL_BLOB_FILE find-mmr-entries --log-tenant $PROD_PUBLIC_TENANT_ID)
    assertEquals "verifying finding the mmr entry of a known public prod event from a local log should match mmr index 663" "matches: [663]" "$output"
  fi
}
