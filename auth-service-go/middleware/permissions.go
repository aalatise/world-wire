package middleware

// Permissions : defined permissions for middleware
func EndPointPermissions() string {
	return `{
	"permissions": {
		"jwt": {
			"default": {
				"method": {
					"GET": {
						"endpoint": {
							"/v1/client/message": {
								"role": {
									"allow": true
								}
							},
							"/v1/anchor/address": {
								"role": {
									"allow": true
								}
							},
							"/v1/client/accounts": {
								"role": {
									"allow": true
								}
							},
							"/v1/client/accounts/{account_name}": {
								"role": {
									"allow": true
								}
							},
							"/v1/client/assets": {
								"role": {
									"allow": true
								}
							},
							"/v1/client/assets/accounts/{account_name}": {
								"role": {
									"allow": true
								}
							},
							"/v1/client/assets/issued": {
								"role": {
									"allow": true
								}
							},
							"/v1/client/assets/participants/{participant_id}": {
								"role": {
									"allow": true
								}
							},
							"/v1/client/balances/accounts/{account_name}": {
								"role": {
									"allow": true
								}
							},
							"/v1/client/obligations": {
								"role": {
									"allow": true
								}
							},
							"/v1/client/obligations/{asset_code}": {
								"role": {
									"allow": true
								}
							},
							"/v1/client/participants/whitelist": {
								"role": {
									"allow": true
								}
							},
							"/v1/client/participants/{participant_id}": {
								"role": {
									"allow": true
								}
							},
							"/v1/client/payout": {
								"role": {
									"allow": true
								}
							},
							"/v1/client/quotes": {
								"role": {
									"allow": true
								}
							},
							"/v1/client/quotes/request/{request_id}": {
								"role": {
									"allow": true
								}
							},
							"/v1/client/quotes/{quote_id}": {
								"role": {
									"allow": true
								}
							},
							"/v1/client/transactions": {
								"role": {
									"allow": true
								}
							},
							"/v1/anchor/fundings/instruction": {
								"role": {
									"allow": true
								}
							},
							"/v1/anchor/fundings/send": {
								"role": {
									"allow": true
								}
							},
							"/v1/anchor/assets/issued/{anchor_id}": {
								"role": {
									"allow": true
								}
							},
							"/v1/anchor/participants": {
								"role": {
									"allow": true
								}
							},
							"/v1/anchor/participants/{participant_id}": {
								"role": {
									"allow": true
								}
							},
							"/v1/anchor/service_check": {
								"role": {
									"allow": true
								}
							},
							"/v1/client/service_check": {
								"role": {
									"allow": true
								}
							},
							"/v1/client/participants/whitelist/object": {
								"role": {
									"allow": true
								}
							},
							"/v1/client/participants": {
								"role": {
									"allow": true
								}
							}
						}
					},
					"POST": {
						"endpoint": {
							"/v1/client/token/refresh": {
								"role": {
									"allow": true
								}
							},
							"/v1/client/accounts/{account_name}/{cursor}": {
								"role": {
									"allow": true
								}
							},
							"/v1/client/accounts/{account_name}/sweep": {
								"role": {
									"allow": true
								}
							},
							"/v1/client/exchange": {
								"role": {
									"allow": true
								}
							},
							"/v1/client/assets": {
								"role": {
									"allow": true
								}
							},
							"/v1/client/fees/request/{participant_id}": {
								"role": {
									"allow": true
								}
							},
							"/v1/client/fees/response/{participant_id}": {
								"role": {
									"allow": true
								}
							},
							"/v1/client/participants": {
								"role": {
									"allow": true
								}
							},
							"/v1/client/quotes": {
								"role": {
									"allow": true
								}
							},
							"/v1/client/quotes/request": {
								"role": {
									"allow": true
								}
							},
							"/v1/client/quotes/{quote_id}": {
								"role": {
									"allow": true
								}
							},
							"/v1/client/sign": {
								"role": {
									"allow": true
								}
							},
							"/v1/client/transactions/reply": {
								"role": {
									"allow": true
								}
							},
							"/v1/client/transactions/send": {
								"role": {
									"allow": true
								}
							},
							"/v1/client/transactions/settle/do": {
								"role": {
									"allow": true
								}
							},
							"/v1/anchor/fundings/instruction": {
								"role": {
									"allow": true
								}
							},
							"/v1/anchor/fundings/send": {
								"role": {
									"allow": true
								}
							},
							"/v1/client/trust": {
								"role": {
									"allow": true
								}
							},
							"/v1/client/participants/whitelist": {
								"role": {
									"allow": true
								}
							},
							"/v1/client/payload/sign": {
								"role": {
									"allow": true
								}
							},
							"/v1/client/transactions/redeem": {
								"role": {
									"allow": true
								}
							},
							"/v1/anchor/trust": {
								"role": {
									"allow": true
								}
							},
							"/v1/anchor/trust/{anchor_id}": {
								"role": {
									"allow": true
								}
							},
							"/v1/anchor/assets/redeem": {
								"role": {
									"allow": true
								}
							}
						}
					},
					"PUT": {
						"endpoint": null
					},
					"DELETE": {
						"endpoint": {
							"/v1/client/quotes": {
								"role": {
									"allow": true
								}
							},
							"/v1/client/quotes/{quote_id}": {
								"role": {
									"allow": true
								}
							},
							"/v1/client/participants/whitelist": {
								"role": {
									"allow": true
								}
							}
						}
					}
				}
			}
		},
		"participant_permissions": {
			"default": {
				"method": {
					"GET": {
						"endpoint": {
							"/v1/client/message": {
								"role": {
									"admin": true,
									"manager": true
								}
							},
							"/v1/anchor/participants": {
								"role": {
									"admin": true,
									"manager": true,
									"viewer": true
								}
							},
							"/v1/anchor/address": {
								"role": {
									"admin": true,
									"manager": true
								}
							},
							"/v1/client/accounts": {
								"role": {
									"admin": true,
									"manager": true
								}
							},
							"/v1/client/accounts/{account_name}": {
								"role": {
									"admin": true,
									"manager": true
								}
							},
							"/v1/client/assets": {
								"role": {
									"admin": true,
									"manager": true
								}
							},
							"/v1/client/assets/accounts/{account_name}": {
								"role": {
									"admin": true,
									"manager": true
								}
							},
							"/v1/client/assets/issued": {
								"role": {
									"admin": true,
									"manager": true
								}
							},
							"/v1/client/assets/participants/{participant_id}": {
								"role": {
									"admin": true,
									"manager": true
								}
							},
							"/v1/client/balances/accounts/{account_name}": {
								"role": {
									"admin": true,
									"manager": true
								}
							},
							"/v1/client/obligations": {
								"role": {
									"admin": true,
									"manager": true
								}
							},
							"/v1/client/obligations/{asset_code}": {
								"role": {
									"admin": true,
									"manager": true
								}
							},
							"/v1/client/participants/whitelist": {
								"role": {
									"admin": true,
									"manager": true
								}
							},
							"/v1/client/participants/{participant_id}": {
								"role": {
									"admin": true,
									"manager": true
								}
							},
							"/v1/client/transactions": {
								"role": {
									"admin": true,
									"manager": true
								}
							},
							"/v1/anchor/service_check": {
								"role": {
									"admin": true,
									"manager": true
								}
							},
							"/v1/anchor/participants/{participant_id}": {
								"role": {
									"admin": true,
									"manager": true
								}
							},
							"/v1/anchor/assets/issued/{anchor_id}": {
								"role": {
									"admin": true,
									"manager": true
								}
							},
							"/v1/client/service_check": {
								"role": {
									"admin": true,
									"manager": true
								}
							},
							"/v1/client/participants/whitelist/object": {
								"role": {
									"admin": true,
									"manager": true
								}
							},
							"/v1/client/participants": {
								"role": {
									"admin": true,
									"manager": true
								}
							}
						}
					},
					"POST": {
						"endpoint": null
					},
					"PUT": {
						"endpoint": null
					},
					"DELETE": {
						"endpoint": null
					}
				}
			},
			"maker_checker": {
				"method": {
					"GET": {
						"endpoint": null
					},
					"POST": {
						"endpoint": {
							"/v1/client/assets": {
								"role": {
									"admin": true,
									"manager": true
								}
							},
							"/v1/anchor/fundings/instruction": {
								"role": {
									"admin": true,
									"manager": true
								}
							},
							"/v1/anchor/fundings/send": {
								"role": {
									"admin": true,
									"manager": true
								}
							},
							"/v1/anchor/trust/{anchor_id}": {
								"role": {
									"admin": true,
									"manager": true
								}
							},
							"/v1/client/trust": {
								"role": {
									"admin": true,
									"manager": true
								}
							},
							"/v1/client/participants/whitelist": {
								"role": {
									"admin": true,
									"manager": true
								}
							},
							"/v1/client/transactions/settle/da": {
								"role": {
									"admin": true,
									"manager": true
								}
							},
							"/v1/client/accounts/{account_name}/sweep": {
								"role": {
									"admin": true,
									"manager": true
								}
							}
						}
					},
					"PUT": {
						"endpoint": null
					},
					"DELETE": {
						"endpoint": {
							"/v1/client/participants/whitelist": {
								"role": {
									"admin": true,
									"manager": true
								}
							}
						}
					}
				}
			}
		},
		"super_permissions": {
			"default": {
				"method": {
					"GET": {
						"endpoint": {
							"/v1/admin/anchor/{anchor_domain}/onboard/assets": {
								"role": {
									"admin": true,
									"manager": true
								}
							},
							"/v1/admin/anchor/{anchor_domain}/register": {
								"role": {
									"admin": true,
									"manager": true
								}
							},
							"/v1/admin/anchor/assets/issued/{anchor_id}": {
								"role": {
									"admin": true,
									"manager": true
								}
							},
							"/v1/admin/blocklist": {
								"role": {
									"admin": true,
									"manager": true
								}
							},
							"/v1/admin/pr": {
								"role": {
									"admin": true,
									"manager": true
								}
							},
							"/v1/admin/pr/domain/{participant_id}": {
								"role": {
									"admin": true,
									"manager": true
								}
							},
							"/v1/onboarding/accounts/{account_name}": {
								"role": {
									"admin": true,
									"manager": true
								}
							},
							"/v1/registry/participants": {
								"role": {
									"admin": true,
									"manager": true
								}
							},
							"/v1/admin/payout": {
								"role": {
									"admin": true,
									"manager": true
								}
							},
							"/v1/admin/service_check": {
								"role": {
									"admin": true,
									"manager": true
								}
							}
						}
					},
					"POST": {
						"endpoint": {
							"/v1/admin/payout": {
								"role": {
									"admin": true,
									"manager": true
								}
							},
							"/v1/admin/payout/csv": {
								"role": {
									"admin": true,
									"manager": true
								}
							},
							"/v1/admin/pr": {
								"role": {
									"admin": true,
									"manager": true
								}
							}
						}
					},
					"PATCH": {
						"endpoint": {
							"/v1/admin/payout": {
								"role": {
									"admin": true,
									"manager": true
								}
							}
						}
					},
					"PUT": {
						"endpoint": {
							"/v1/admin/pr/{participant_id}/status": {
								"role": {
									"admin": true,
									"manager": true
								}
							},
							"/v1/admin/pr/{participant_id}": {
								"role": {
									"admin": true,
									"manager": true
								}
							}
						}
					},
					"DELETE": {
						"endpoint": {
							"/v1/admin/payout": {
								"role": {
									"admin": true,
									"manager": true
								}
							}
						}
					}
				}
			},
			"maker_checker": {
				"method": {
					"GET": {
						"endpoint": null
					},
					"POST": {
						"endpoint": {
							"/v1/admin/anchor/{anchor_domain}/onboard/assets": {
								"role": {
									"admin": true,
									"manager": true
								}
							},
							"/v1/admin/anchor/{anchor_domain}/register": {
								"role": {
									"admin": true,
									"manager": true
								}
							},
							"/v1/onboarding/accounts/{account_name}": {
								"role": {
									"admin": true,
									"manager": true
								}
							},
							"/v1/admin/accounts/{account_name}": {
								"role": {
									"admin": true,
									"manager": true
								}
							},
							"/v1/deploy/participant": {
								"role": {
									"admin": true,
									"manager": true
								}
							},
							"/v1/admin/blocklist": {
								"role": {
									"admin": true,
									"manager": true
								}
							},
							"/v1/admin/suspend/{participant_id}/{account_name}": {
								"role": {
									"admin": true,
									"manager": true
								}
							},
							"/v1/admin/reactivate/{participant_id}/{account_name}": {
								"role": {
									"admin": true,
									"manager": true
								}
							},
							"/v1/admin/transaction": {
								"role": {
									"admin": true,
									"manager": true
								}
							}
						}
					},
					"PUT": {
						"endpoint": null
					},
					"DELETE": {
						"endpoint": {
							"/v1/admin/blocklist": {
								"role": {
									"admin": true,
									"manager": true
								}
							}
						}
					}
				}
			}
		}
	}
}`

}