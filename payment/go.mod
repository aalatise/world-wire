module github.com/IBM/world-wire/utility/payment

go 1.16

replace (
	github.com/IBM/world-wire/gftn-models => ./../gftn-models
	github.com/IBM/world-wire/iso20022 => ./../iso20022
	github.com/IBM/world-wire/utility/common => ../../../../../../github.com/IBM/world-wire/utility/common
	github.com/IBM/world-wire/utility/global-environment => ./../utility/global-environment
)
