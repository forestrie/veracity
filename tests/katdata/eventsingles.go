package katdata

var (
	KnownGoodPublicEvent = []byte(`{
  "identity": "publicassets/87dd2e5a-42b4-49a5-8693-97f40a5af7f8/events/a022f458-8e55-4d63-a200-4172a42fc2aa",
  "asset_identity": "publicassets/87dd2e5a-42b4-49a5-8693-97f40a5af7f8",
  "event_attributes": {
    "arc_access_policy_always_read": [
      {
        "wallet": "0x0E29670b420B7f2E8E699647b632cdE49D868dA7",
        "tessera": "SmL4PHAHXLdpkj/c6Xs+2br+hxqLmhcRk75Hkj5DyEQ="
      }
    ],
    "arc_access_policy_asset_attributes_read": [
      {
        "attribute": "*",
        "0x4609ea6bbe85F61bc64760273ce6D89A632B569f": "wallet",
        "SmL4PHAHXLdpkj/c6Xs+2br+hxqLmhcRk75Hkj5DyEQ=": "tessera"
      }
    ],
    "arc_access_policy_event_arc_display_type_read": [
      {
        "value": "*",
        "0x4609ea6bbe85F61bc64760273ce6D89A632B569f": "wallet",
        "SmL4PHAHXLdpkj/c6Xs+2br+hxqLmhcRk75Hkj5DyEQ=": "tessera"
      }
    ]
  },
  "asset_attributes": {
    "arc_display_type": "public-test",
    "arc_display_name": "Dava Derby"
  },
  "operation": "NewAsset",
  "behaviour": "AssetCreator",
  "timestamp_declared": "2024-05-24T07:26:58Z",
  "timestamp_accepted": "2024-05-24T07:26:58Z",
  "timestamp_committed": "2024-05-24T07:27:00.200Z",
  "principal_declared": {
    "issuer": "",
    "subject": "",
    "display_name": "",
    "email": ""
  },
  "principal_accepted": {
    "issuer": "",
    "subject": "",
    "display_name": "",
    "email": ""
  },
  "confirmation_status": "CONFIRMED",
  "transaction_id": "0xc891533b1806555fff9ab853cd9ce1bb2c00753609070a875a44ec53a6c1213b",
  "block_number": 7932,
  "transaction_index": 1,
  "from": "0x0E29670b420B7f2E8E699647b632cdE49D868dA7",
  "tenant_identity": "tenant/7dfaa5ef-226f-4f40-90a5-c015e59998a8",
  "merklelog_entry": {
    "commit": {
      "index": "663",
      "idtimestamp": "018fa97ef269039b00"
    },
    "confirm": {
      "mmr_size": "664",
      "root": "/rlMNJhlay9CUuO3LgX4lSSDK6dDhtKesCO50CtrHr4=",
      "timestamp": "1716535620409",
      "idtimestamp": "",
      "signed_tree_head": ""
    },
    "unequivocal": null
  }
}`)

	KnownGoodEventsV1Event = []byte(`{
    "identity": "events/0194b168-bac0-75e6-bbc4-a47cc45bdbf5",
    "attributes": {
        "4": "finally add in the eggs",
        "5": "put in the over until golden brown",
        "1": "pour flour and milk into bowl",
        "2": "mix together until gloopy",
        "3": "slowly add in the sugar while still mixing"
    },
    "trails": [
        "cake"
    ],
    "origin_tenant": "tenant/97e90a09-8c56-40df-a4de-42fde462ef6f",
    "created_by": "a3732a3f-1406-45b6-bdce-2976945752fc",
    "created_at": 1738143218368,
    "confirmation_status": "CONFIRMED",
    "ledger_entry": {
        "index": "4",
        "idtimestamp": "0194b168bcde03be00",
        "log_id": "97e90a09-8c56-40df-a4de-42fde462ef6f"
    }
}`)

	// The 'tamper' is the 'a' from a single arc_ event attribute has been clipped.
	KnownTamperedPublicEvent = []byte(`{
  "identity": "publicassets/87dd2e5a-42b4-49a5-8693-97f40a5af7f8/events/a022f458-8e55-4d63-a200-4172a42fc2aa",
  "asset_identity": "publicassets/87dd2e5a-42b4-49a5-8693-97f40a5af7f8",
  "event_attributes": {
    "rc_access_policy_always_read": [
      {
        "wallet": "0x0E29670b420B7f2E8E699647b632cdE49D868dA7",
        "tessera": "SmL4PHAHXLdpkj/c6Xs+2br+hxqLmhcRk75Hkj5DyEQ="
      }
    ],
    "arc_access_policy_asset_attributes_read": [
      {
        "attribute": "*",
        "0x4609ea6bbe85F61bc64760273ce6D89A632B569f": "wallet",
        "SmL4PHAHXLdpkj/c6Xs+2br+hxqLmhcRk75Hkj5DyEQ=": "tessera"
      }
    ],
    "arc_access_policy_event_arc_display_type_read": [
      {
        "value": "*",
        "0x4609ea6bbe85F61bc64760273ce6D89A632B569f": "wallet",
        "SmL4PHAHXLdpkj/c6Xs+2br+hxqLmhcRk75Hkj5DyEQ=": "tessera"
      }
    ]
  },
  "asset_attributes": {
    "arc_display_type": "public-test",
    "arc_display_name": "Dava Derby"
  },
  "operation": "NewAsset",
  "behaviour": "AssetCreator",
  "timestamp_declared": "2024-05-24T07:26:58Z",
  "timestamp_accepted": "2024-05-24T07:26:58Z",
  "timestamp_committed": "2024-05-24T07:27:00.200Z",
  "principal_declared": {
    "issuer": "",
    "subject": "",
    "display_name": "",
    "email": ""
  },
  "principal_accepted": {
    "issuer": "",
    "subject": "",
    "display_name": "",
    "email": ""
  },
  "confirmation_status": "CONFIRMED",
  "transaction_id": "0xc891533b1806555fff9ab853cd9ce1bb2c00753609070a875a44ec53a6c1213b",
  "block_number": 7932,
  "transaction_index": 1,
  "from": "0x0E29670b420B7f2E8E699647b632cdE49D868dA7",
  "tenant_identity": "tenant/7dfaa5ef-226f-4f40-90a5-c015e59998a8",
  "merklelog_entry": {
    "commit": {
      "index": "663",
      "idtimestamp": "018fa97ef269039b00"
    },
    "confirm": {
      "mmr_size": "664",
      "root": "/rlMNJhlay9CUuO3LgX4lSSDK6dDhtKesCO50CtrHr4=",
      "timestamp": "1716535620409",
      "idtimestamp": "",
      "signed_tree_head": ""
    },
    "unequivocal": null
  }
}`)
)
