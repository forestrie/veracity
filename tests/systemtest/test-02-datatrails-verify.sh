testVerifySingleEvent() {
  echo "disabled, datatrails public events service has been shutdown"
  if false; then
    # Check if the response status code is 200
    local response
    response=$(curl -sL -w "%{http_code}" $DATATRAILS_URL/archivist/v2/$PUBLIC_EVENT_ID -o /dev/null)
    assertEquals 200 "$response"
    # Verify the event and check if the exit code is 0
    curl -sL $DATATRAILS_URL/archivist/v2/$PUBLIC_EVENT_ID | $VERACITY_INSTALL --data-url $DATATRAILS_URL/verifiabledata --tenant=$PROD_PUBLIC_TENANT_ID verify-included
    assertEquals "Verifying a valid single event should result in a 0 exit code" 0 $?
  fi
}

testVerifyListEvents() {
  echo "disabled, datatrails public events service has been shutdown"
  if false; then

    # Check if the response status code is 200
    local response
    response=$(curl -sL -w "%{http_code}" $DATATRAILS_URL/archivist/v2/$PUBLIC_ASSET_ID/events -o /dev/null)
    assertEquals 200 "$response"
    # Verify the events on the asset and check if the exit code is 0
    curl -sL $DATATRAILS_URL/archivist/v2/$PUBLIC_ASSET_ID/events | $VERACITY_INSTALL --data-url $DATATRAILS_URL/verifiabledata --tenant=$PROD_PUBLIC_TENANT_ID verify-included
    assertEquals "Verifying events on a vaid asset should result in a 0 exit code" 0 $?
  fi
}

testVerifySingleEventWithLocalMassifCopy() {
  echo "disabled, datatrails public events service has been shutdown"
  if false; then

    # Verify the event and check if the exit code is 0
    curl -sL $DATATRAILS_URL/archivist/v2/$PUBLIC_EVENT_ID | $VERACITY_INSTALL --data-local $PROD_LOCAL_BLOB_FILE --tenant=$PROD_PUBLIC_TENANT_ID verify-included
    assertEquals "verifying valid events with a local copy of the massif should result in a 0 exit code" 0 $?
  fi
}
