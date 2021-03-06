package parse

import (
	"fmt"
	"testing"
)

func TestValidateSchema(t *testing.T) {
	for cnt := 1; cnt < 10; cnt++ {
		err := ValidateSchema(`<?xml version="1.0" encoding="UTF-8"?>
											<Message xmlns="urn:worldwire" xmlns:cct="urn:iso:std:iso:20022:tech:xsd:pacs.008.001.07" xmlns:head="urn:iso:std:iso:20022:tech:xsd:head.001.001.01">
												<AppHdr>
													<head:Fr>
														<head:FIId>
															<head:FinInstnId>
																<head:BICFI>SGPTTEST003</head:BICFI>
																<head:Othr>
																	<head:Id>testparticipant3dev</head:Id>
																</head:Othr>
															</head:FinInstnId>
														</head:FIId>
													</head:Fr>
													<head:To>
														<head:FIId>
															<head:FinInstnId>
																<head:BICFI>WORLDWIRE00</head:BICFI>
																<head:Othr>
																	<head:Id>WW</head:Id>
																</head:Othr>
															</head:FinInstnId>
														</head:FIId>
													</head:To>
													<head:BizMsgIdr>B20190819SGPTTEST003BAA4710449</head:BizMsgIdr>
													<head:MsgDefIdr>pacs.008.001.07</head:MsgDefIdr>
													<head:CreDt>2019-08-19T13:12:18Z</head:CreDt>
												<Sgntr xmlns="urn:iso:std:iso:20022:tech:xsd:head.001.001.01"><Signature xmlns="http://www.w3.org/2000/09/xmldsig#"><SignedInfo xmlns="http://www.w3.org/2000/09/xmldsig#"><CanonicalizationMethod xmlns="http://www.w3.org/2000/09/xmldsig#" Algorithm="http://www.w3.org/2001/10/xml-exc-c14n#"/><SignatureMethod xmlns="http://www.w3.org/2000/09/xmldsig#" Algorithm="http://www.w3.org/2009/xmldsig11#rsa-sha256"/><Reference xmlns="http://www.w3.org/2000/09/xmldsig#"><Transforms xmlns="http://www.w3.org/2000/09/xmldsig#"><Transform xmlns="http://www.w3.org/2000/09/xmldsig#" Algorithm="http://www.w3.org/2000/09/xmldsig#enveloped-signature"/><Transform xmlns="http://www.w3.org/2000/09/xmldsig#" Algorithm="http://www.w3.org/2001/10/xml-exc-c14n#"/></Transforms><DigestMethod xmlns="http://www.w3.org/2000/09/xmldsig#" Algorithm="http://www.w3.org/2001/04/xmlenc#sha256"/><DigestValue xmlns="http://www.w3.org/2000/09/xmldsig#">ogZVF9q472RM/4kCbZcee3uG0z8xBi7fyapdzvzxbi0=</DigestValue></Reference></SignedInfo><SignatureValue xmlns="http://www.w3.org/2000/09/xmldsig#">qHPCstyKGhTviY/ZGCHELEj8/T9Fn2v//Vtd/tm5nIimsfVypyPOMvPzMAZyZJfgjfX4LMkPPDL4H8IHAu9bBA==</SignatureValue><KeyInfo xmlns="http://www.w3.org/2000/09/xmldsig#"><X509Data xmlns="http://www.w3.org/2000/09/xmldsig#"><X509Certificate xmlns="http://www.w3.org/2000/09/xmldsig#">GD3YO32Q2RKC5VBBQR5X5QFP5K6BOJT4NW7QD2I5KGXL24B7CIH77WA7</X509Certificate></X509Data></KeyInfo></Signature></Sgntr></AppHdr>
												<FIToFICstmrCdtTrf>
													<cct:GrpHdr>
														<cct:MsgId>USDDO19082019SGPTTEST00377793380333</cct:MsgId>
														<cct:CreDtTm>2019-08-19T13:12:18</cct:CreDtTm>
														<cct:NbOfTxs>1</cct:NbOfTxs>
														<cct:SttlmInf>
															<cct:SttlmMtd>WWDO</cct:SttlmMtd>
															<cct:SttlmAcct>
																<cct:Id>
																	<cct:Othr>
																		<cct:Id>testparticipant3dev</cct:Id>
																	</cct:Othr>
																</cct:Id>
																<cct:Nm>issuing</cct:Nm>
															</cct:SttlmAcct>
														</cct:SttlmInf>
														<cct:PmtTpInf>
															<cct:SvcLvl>
																<cct:Prtry>testparticipant3dev</cct:Prtry>
															</cct:SvcLvl>
														</cct:PmtTpInf>
														<cct:InstgAgt>
															<cct:FinInstnId>
																<cct:BICFI>SGPTTEST003</cct:BICFI>
																<cct:Othr>
																	<cct:Id>testparticipant3dev</cct:Id>
																</cct:Othr>
															</cct:FinInstnId>
														</cct:InstgAgt>
														<cct:InstdAgt>
															<cct:FinInstnId>
																<cct:BICFI>SGPTTEST004</cct:BICFI>
																<cct:Othr>
																	<cct:Id>testparticipant4dev</cct:Id>
																</cct:Othr>
															</cct:FinInstnId>
														</cct:InstdAgt>
													</cct:GrpHdr>
													<cct:CdtTrfTxInf>
														<cct:PmtId>
															<cct:InstrId>USDDO20190819SGPTTEST003B3889747112</cct:InstrId>
															<cct:EndToEndId>USDDO19082019SGPTTEST00377793380333</cct:EndToEndId>
															<cct:TxId>USDDO19082019SGPTTEST00377793380333</cct:TxId>
														</cct:PmtId>
														<cct:IntrBkSttlmAmt Ccy="USDDO">0.02</cct:IntrBkSttlmAmt>
														<cct:IntrBkSttlmDt>2019-08-19</cct:IntrBkSttlmDt>
														<cct:InstdAmt Ccy="USDDO">0.02</cct:InstdAmt>
														<cct:XchgRate>1</cct:XchgRate>
														<cct:ChrgBr>DEBT</cct:ChrgBr>
														<cct:ChrgsInf>
															<cct:Amt Ccy="USDDO">0</cct:Amt>
															<cct:Agt>
																<cct:FinInstnId>
																	<cct:BICFI>SGPTTEST004</cct:BICFI>
																	<cct:Othr>
																		<cct:Id>testparticipant3dev</cct:Id>
																	</cct:Othr>
																</cct:FinInstnId>
															</cct:Agt>
														</cct:ChrgsInf>
														<cct:Dbtr>
															<cct:Nm>ABC</cct:Nm>
															<cct:PstlAdr>
																<cct:StrtNm>Times Square</cct:StrtNm>
																<cct:BldgNb>7</cct:BldgNb>
																<cct:PstCd>NY 10036</cct:PstCd>
																<cct:TwnNm>New York</cct:TwnNm>
																<cct:Ctry>US</cct:Ctry>
															</cct:PstlAdr>
														</cct:Dbtr>
														<cct:DbtrAgt>
															<cct:FinInstnId>
																<cct:BICFI>SGPTTEST003</cct:BICFI>
																<cct:Othr>
																	<cct:Id>testparticipant3dev</cct:Id>
																</cct:Othr>
															</cct:FinInstnId>
														</cct:DbtrAgt>
														<cct:CdtrAgt>
															<cct:FinInstnId>
																<cct:BICFI>SGPTTEST004</cct:BICFI>
																<cct:Othr>
																	<cct:Id>testparticipant4dev</cct:Id>
																</cct:Othr>
															</cct:FinInstnId>
														</cct:CdtrAgt>
														<cct:Cdtr>
															<cct:Nm>DEF</cct:Nm>
															<cct:PstlAdr>
																<cct:StrtNm>Mark Lane</cct:StrtNm>
																<cct:BldgNb>55</cct:BldgNb>
																<cct:PstCd>EC3R7NE</cct:PstCd>
																<cct:TwnNm>London</cct:TwnNm>
																<cct:Ctry>GB</cct:Ctry>
																<cct:AdrLine>Corn Exchange 5th Floor</cct:AdrLine>
															</cct:PstlAdr>
														</cct:Cdtr>
														<cct:SplmtryData>
															<cct:PlcAndNm>payout</cct:PlcAndNm>
															<cct:Envlp>
																<cct:Id>dqLUvpwLma5py2zBtGM0uWeWsJkzNT</cct:Id>
															</cct:Envlp>
														</cct:SplmtryData>
														<cct:SplmtryData>
															<cct:PlcAndNm>fee</cct:PlcAndNm>
															<cct:Envlp>
																<cct:Id>23982398</cct:Id>
															</cct:Envlp>
														</cct:SplmtryData>
													</cct:CdtTrfTxInf>
												</FIToFICstmrCdtTrf>
											</Message>`)
		fmt.Println("Error:", err)
	}
}
