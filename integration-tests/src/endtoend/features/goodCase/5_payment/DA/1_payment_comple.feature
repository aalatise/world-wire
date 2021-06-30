
Feature: World Wire goodCase/4_payment/DA/1_payment_comple
  In order to helping a customer send money
  As a developer
  I want to send payment in Worldwire

  Scenario Outline: 1_OFI_get_fee_send_asset_to_RFI
    Given id: <participant_id> check participant: <RFI_participant_id>, was in Workd Wire, apiURL: <participant_api_url>
    When  id: <participant_id> check the asset_code: <sending_asset_code> asset_type: <sending_asset_type> issuer: <sending_asset_issuer> already issue by worldwire, apiURL:<participant_api_url>
    And id: <participant_id> add participant: <RFI_participant_id> - whitelistURL: <global_whitelist_url>
    And id: <participant_id> check participant: <RFI_participant_id> was in the whitelist - whitelistURL: <global_whitelist_url>
    And id: <RFI_participant_id> add participant: <participant_id> - whitelistURL: <global_whitelist_url>
    And id: <RFI_participant_id> check participant: <participant_id> was in the whitelist - whitelistURL: <global_whitelist_url>
    And id: <participant_id> query payout area by <payout_query> - payoutURL: <global_payout_url>
    # Then id:  <participant_id> get the fee with fee_request_id <fee_request_id> receiver <RFI_participant_id> sending_asset_code <sending_asset_code> sending_asset_type <sending_asset_type> sending_asset_issuer_id <sending_asset_issuer> amount_gross <sending_amount> asset_payout <asset_payout>, feeURL:<participant_fee_url>
    # ########### FOR NEW CALLBACK TESTING CASE USE ###########
    Then id:  <participant_id> request fee with fee_request_id <fee_request_id> receiver <RFI_participant_id> sending_asset_code <sending_asset_code> sending_asset_type <sending_asset_type> sending_asset_issuer_id <sending_asset_issuer> amount_gross <sending_amount> asset_payout <asset_payout>, feeURL:<participant_fee_url>
    And id: <RFI_participant_id> pick up message from RFI_FEE topic sent by ofi <participant_id> should get fee_request_id <fee_request_id> receiver <RFI_participant_id> sending_asset_code <sending_asset_code> sending_asset_type <sending_asset_type> sending_asset_issuer_id <sending_asset_issuer> amount_gross <sending_amount> asset_payout <asset_payout>, wwGatewayURL: <rfi_wwGatewayURL>
    And id: <RFI_participant_id> caculate fee amount as: <fee_amount>, fee amount_payout as: <sending_amount> asset_payout <asset_payout> amount_settlement <sending_amount> sending_asset_code <sending_asset_code> sending_asset_type <sending_asset_type> sending_asset_issuer_id <sending_asset_issuer> and response fee message to OFI: <participant_id> OFI_FEE topic, fee: <participant_fee_url>
    And id: <participant_id> pick up message from OFI_FEE topic should get fee amount as: <fee_amount>, fee amount_payout as: <sending_amount> asset_payout <asset_payout> amount_settlement <sending_amount> sending_asset_code <sending_asset_code> sending_asset_type <sending_asset_type> sending_asset_issuer_id <sending_asset_issuer>, wwGatewayURL: <ofi_wwGatewayURL>
    # ########### FOR NEW CALLBACK TESTING CASE USE ###########
    Then id: <participant_id> send asset from sender_bic <sender_bic> sending_account_name <sending_account_name> sending_bank_name <sending_bank_name> sending_street_name <sending_street_name> sending_building_number <sending_building_number> sending_post_code <sending_post_code> sending_town_name <sending_town_name> sending_country <sending_country> with settlement_method <settlement_method> asset_code <sending_asset_code> asset_issuer <asset_issuer_participant_id> charger_bic <charger_bic> to receiver <receiver_id> recever_bic <receiver_bic> receiver_bank_name <receiver_bank_name> receiver_street_name <receiver_street_name> receiver_building_number <receiver_building_number> receiver_post_code <receiver_post_code> receiver_town_name <receiver_town_name> receiver_country <receiver_country> receiver_address_line <receiver_address_line>, sendURL: <send_service_url>

    Examples:
      | participant_id             | participant_api_url             | RFI_participant_id         | asset_issuer_participant_id | sending_asset_code          | sending_asset_type          | sending_asset_issuer | global_whitelist_url           | fee_request_id            | amount_payout            | asset_payout                | participant_fee_url             | send_service_url                 | sender_bic                  | sending_account_name                           | sending_bank_name                 | sending_street_name                      | sending_building_number                      | sending_post_code                      | sending_town_name                      | sending_country                      | settlement_method              | sending_amount                         | charger_bic                 | receiver_id                | receiver_bic                | receiver_bank_name                | receiver_street_name                     | receiver_building_number                     | receiver_post_code                     | receiver_town_name                     | receiver_country                     | receiver_address_line                     | payout_query           | global_payout_url                  | ofi_wwGatewayURL                      | rfi_wwGatewayURL                      | fee_amount                                 |
      | "ENV_KEY_PARTICIPANT_1_ID" | "ENV_KEY_PARTICIPANT_1_API_URL" | "ENV_KEY_PARTICIPANT_2_ID" | "ENV_KEY_ANCHOR_ID"         | "ENV_KEY_ANCHOR_ASSET_CODE" | "ENV_KEY_ANCHOR_ASSET_TYPE" | "ENV_KEY_ANCHOR_ID"  | "ENV_KEY_PARTICIPANT_1_WL_URL" | "ENV_KEY_FEE_REQUEST_ID1" | "ENV_KEY_PAYOUT_AMOUNT1" | "ENV_KEY_PAYOUT_ASSET1" | "ENV_KEY_PARTICIPANT_1_FEE_URL" | "ENV_KEY_PARTICIPANT_1_SEND_URL" | "ENV_KEY_PARTICIPANT_1_BIC" | "ENV_KEY_PARTICIPANT_1_OPERATING_ACCOUNT_NAME" | "ENV_KEY_PARTICIPANT_1_BANK_NAME" | "ENV_KEY_PARTICIPANT_1_BANK_STREET_NAME" | "ENV_KEY_PARTICIPANT_1_BANK_BUILDING_NUMBER" | "ENV_KEY_PARTICIPANT_1_BANK_POST_CODE" | "ENV_KEY_PARTICIPANT_1_BANK_TOWN_NAME" | "ENV_KEY_PARTICIPANT_1_BANK_COUNTRY" | "ENV_KEY_SETTLEMENT_METHOD_DA" | "ENV_KEY_PARTICIPANT_1_SENDING_AMOUNT" | "ENV_KEY_PARTICIPANT_2_BIC" | "ENV_KEY_PARTICIPANT_2_ID" | "ENV_KEY_PARTICIPANT_2_BIC" | "ENV_KEY_PARTICIPANT_2_BANK_NAME" | "ENV_KEY_PARTICIPANT_2_BANK_STREET_NAME" | "ENV_KEY_PARTICIPANT_2_BANK_BUILDING_NUMBER" | "ENV_KEY_PARTICIPANT_2_BANK_POST_CODE" | "ENV_KEY_PARTICIPANT_2_BANK_TOWN_NAME" | "ENV_KEY_PARTICIPANT_2_BANK_COUNTRY" | "ENV_KEY_PARTICIPANT_2_BANK_ADDRESS_LINE" | "ENV_KEY_PAYOUT_QUERY" | "ENV_KEY_PARTICIPANT_1_PAYOUT_URL" | "ENV_KEY_PARTICIPANT_1_WWGATEWAY_URL" | "ENV_KEY_PARTICIPANT_2_WWGATEWAY_URL" | "ENV_KEY_PARTICIPANT_1_SENDING_FEE_AMOUNT" |
  # | "ENV_KEY_PARTICIPANT_2_ID" | "ENV_KEY_PARTICIPANT_2_API_URL"  | "ENV_KEY_PARTICIPANT_1_ID" | "ENV_KEY_ANCHOR_ID"         | "ENV_KEY_ANCHOR_ASSET_CODE"              | "ENV_KEY_ANCHOR_ASSET_TYPE"              | "ENV_KEY_ANCHOR_ID"              | "ENV_KEY_PARTICIPANT_2_WL_URL" | "ENV_KEY_FEE_REQUEST_ID1" | "ENV_KEY_PAYOUT_AMOUNT1"   | "ENV_KEY_ANCHOR_ASSET_CODE"         | "ENV_KEY_PARTICIPANT_2_FEE_URL"   | "ENV_KEY_PARTICIPANT_2_SEND_URL" | "ENV_KEY_PARTICIPANT_2_BIC"         | "ENV_KEY_PARTICIPANT_2_OPERATING_ACCOUNT_NAME" | "ENV_KEY_PARTICIPANT_2_BANK_NAME"| "ENV_KEY_PARTICIPANT_2_BANK_STREET_NAME" | "ENV_KEY_PARTICIPANT_2_BANK_BUILDING_NUMBER" | "ENV_KEY_PARTICIPANT_2_BANK_POST_CODE" | "ENV_KEY_PARTICIPANT_2_BANK_TOWN_NAME" | "ENV_KEY_PARTICIPANT_2_BANK_COUNTRY" | "ENV_KEY_SETTLEMENT_METHOD_DA"  | "ENV_KEY_PARTICIPANT_2_SENDING_AMOUNT" | "ENV_KEY_PARTICIPANT_1_BIC" | "ENV_KEY_PARTICIPANT_1_ID"   | "ENV_KEY_PARTICIPANT_1_BIC"  | "ENV_KEY_PARTICIPANT_1_BANK_NAME"  | "ENV_KEY_PARTICIPANT_1_BANK_STREET_NAME"  | "ENV_KEY_PARTICIPANT_1_BANK_BUILDING_NUMBER"  | "ENV_KEY_PARTICIPANT_1_BANK_POST_CODE"  | "ENV_KEY_PARTICIPANT_1_BANK_TOWN_NAME"   | "ENV_KEY_PARTICIPANT_1_BANK_COUNTRY"   | "ENV_KEY_PARTICIPANT_1_BANK_ADDRESS_LINE" | "ENV_KEY_PAYOUT_QUERY" | "ENV_KEY_PARTICIPANT_2_PAYOUT_URL" | "ENV_KEY_PARTICIPANT_2_WWGATEWAY_URL" | "ENV_KEY_PARTICIPANT_1_WWGATEWAY_URL" | "ENV_KEY_PARTICIPANT_2_SENDING_FEE_AMOUNT" |



  Scenario Outline: 2_RFI_receive_asset_from_OFI
    Given id: <participant_id> get account_name: <receive_account_name> address from WW, participant_api_url:<participant_api_url>
    And id: <participant_id> check asset_code: <asset_code> issuer: <asset_issuer> account_name: <receive_account_name> balance before transaction, participant_api_url:<participant_api_url>
    And id: <send_participant_id> check asset_code: <asset_code> issuer: <asset_issuer> account_name: <send_account_name> balance before transaction, participant_api_url:<send_participant_api_url>
    Then id: <participant_id> participant_bic: <participant_bic> finished federation and compliance check response to send_participant: <send_participant_id> send_participant_bic: <send_participant_bic> with federation_status: <federation_status> compliance_status: <compliance_status_1> compliance_status: <compliance_status_2> receive asset_code: <asset_code> account_name: <receive_account_name> settlement_method: <settlement_method>, sendURL: <send_service_url>
    And id: <participant_id> check asset_code: <asset_code> issuer: <asset_issuer> account_name: <receive_account_name> balance increase settle amount ofi: <send_participant_id> participant_api_url:<participant_api_url>


    Examples:
      | participant_id             | participant_api_url             | send_participant_api_url        | send_account_name                              | receive_account_name                           | participant_bic             | send_participant_id        | sending_amount                         | send_participant_bic        | federation_status                         | compliance_status_1                         | compliance_status_2                         | asset_code                  | send_service_url                 | settlement_method              | global_whitelist_url           | asset_issuer        |
      | "ENV_KEY_PARTICIPANT_2_ID" | "ENV_KEY_PARTICIPANT_2_API_URL" | "ENV_KEY_PARTICIPANT_1_API_URL" | "ENV_KEY_PARTICIPANT_1_OPERATING_ACCOUNT_NAME" | "ENV_KEY_PARTICIPANT_2_OPERATING_ACCOUNT_NAME" | "ENV_KEY_PARTICIPANT_2_BIC" | "ENV_KEY_PARTICIPANT_1_ID" | "ENV_KEY_PARTICIPANT_1_SENDING_AMOUNT" | "ENV_KEY_PARTICIPANT_1_BIC" | "ENV_KEY_PARTICIPANT_2_FEDERATION_STATUS" | "ENV_KEY_PARTICIPANT_2_COMPLIANCE_1_STATUS" | "ENV_KEY_PARTICIPANT_2_COMPLIANCE_2_STATUS" | "ENV_KEY_ANCHOR_ASSET_CODE" | "ENV_KEY_PARTICIPANT_2_SEND_URL" | "ENV_KEY_SETTLEMENT_METHOD_DA" | "ENV_KEY_PARTICIPANT_1_WL_URL" | "ENV_KEY_ANCHOR_ID" |
# | "ENV_KEY_PARTICIPANT_1_ID" | "ENV_KEY_PARTICIPANT_1_API_URL"  | "ENV_KEY_PARTICIPANT_2_API_URL" | "ENV_KEY_PARTICIPANT_2_OPERATING_ACCOUNT_NAME" | "ENV_KEY_PARTICIPANT_1_OPERATING_ACCOUNT_NAME" | "ENV_KEY_PARTICIPANT_1_BIC"  | "ENV_KEY_PARTICIPANT_2_ID" | "ENV_KEY_PARTICIPANT_2_SENDING_AMOUNT" | "ENV_KEY_PARTICIPANT_2_BIC" | "ENV_KEY_PARTICIPANT_1_FEDERATION_STATUS" | "ENV_KEY_PARTICIPANT_1_COMPLIANCE_2_STATUS"   | "ENV_KEY_PARTICIPANT_1_COMPLIANCE_1_STATUS" | "ENV_KEY_ANCHOR_ASSET_CODE"         | "ENV_KEY_PARTICIPANT_1_SEND_URL" | "ENV_KEY_SETTLEMENT_METHOD_DA" | "ENV_KEY_PARTICIPANT_2_WL_URL" | "ENV_KEY_ANCHOR_ID" |

