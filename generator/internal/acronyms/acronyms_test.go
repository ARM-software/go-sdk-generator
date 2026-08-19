/*
 * Copyright (C) 2026-2026 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
package acronyms

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadRules(t *testing.T) {
	rules, err := loadRules(context.Background())
	require.NoError(t, err)
	require.NotEmpty(t, rules)
}

func TestApplyKnownAcronyms(t *testing.T) {
	t.Run("replaces configured acronym in pascal case", func(t *testing.T) {
		value, err := ApplyKnownAcronyms("ApiClient")
		require.NoError(t, err)
		assert.Equal(t, "APIClient", value)
	})

	t.Run("replaces consecutive configured acronyms in pascal case", func(t *testing.T) {
		value, err := ApplyKnownAcronyms("OpenAiApiKey")
		require.NoError(t, err)
		assert.Equal(t, "OpenAIAPIKey", value)
	})

	t.Run("replaces configured hal acronym in pascal case", func(t *testing.T) {
		value, err := ApplyKnownAcronyms("HalClient")
		require.NoError(t, err)
		assert.Equal(t, "HALClient", value)
	})

	t.Run("replaces configured cli acronym in pascal case", func(t *testing.T) {
		value, err := ApplyKnownAcronyms("CliTool")
		require.NoError(t, err)
		assert.Equal(t, "CLITool", value)
	})

	t.Run("replaces configured ai ml acronyms in pascal case", func(t *testing.T) {
		//nolint:gosec // identifier-like acronym fixtures such as JwtToken are test inputs, not credentials
		testCases := map[string]string{
			"AiAgent":       "AIAgent",
			"AiopsPipeline": "AIOpsPipeline",
			"LlmProvider":   "LLMProvider",
			"MlModel":       "MLModel",
			"MlopsPlatform": "MLOpsPlatform",
			"SlmRuntime":    "SLMRuntime",
			"NlpPipeline":   "NLPPipeline",
			"CvModel":       "CVModel",
			"AsrModel":      "ASRModel",
			"TtsModel":      "TTSModel",
			"RagPipeline":   "RAGPipeline",
			"McpServer":     "MCPServer",
			"RlPolicy":      "RLPolicy",
			"RlhfTrainer":   "RLHFTrainer",
			"DpoTrainer":    "DPOTrainer",
			"PpoTrainer":    "PPOTrainer",
			"LoraAdapter":   "LoRAAdapter",
			"QloraAdapter":  "QLoRAAdapter",
			"SftDataset":    "SFTDataset",
			"MoeRouter":     "MoERouter",
			"GpuWorker":     "GPUWorker",
			"TpuWorker":     "TPUWorker",
			"CudaKernel":    "CUDAKernel",
			"RopeScaling":   "ROPEScaling",
			"BpeTokenizer":  "BPETokenizer",
			"GptModel":      "GPTModel",
			"BertModel":     "BERTModel",
			"T5Encoder":     "T5Encoder",
			"VaeDecoder":    "VAEDecoder",
			"GanTrainer":    "GANTrainer",
			"CnnEncoder":    "CNNEncoder",
			"RnnDecoder":    "RNNDecoder",
			"LstmLayer":     "LSTMLayer",
			"GruLayer":      "GRULayer",
			"OcrService":    "OCRService",
			"NerService":    "NERService",
			"KnnIndex":      "KNNIndex",
			"SvmClassifier": "SVMClassifier",
			"XgboostModel":  "XGBoostModel",
			"SdkManager":    "SDKManager",
			"ApiClient":     "APIClient",
			"CiPipeline":    "CIPipeline",
			"CdWorkflow":    "CDWorkflow",
			"CliTool":       "CLITool",
			"UiRenderer":    "UIRenderer",
			"TuiRenderer":   "TUIRenderer",
			"UxReview":      "UXReview",
			"ReplSession":   "REPLSession",
			"StdinReader":   "STDINReader",
			"StdoutWriter":  "STDOUTWriter",
			"StderrWriter":  "STDERRWriter",
			"HttpClient":    "HTTPClient",
			"HttpsProxy":    "HTTPSProxy",
			"JsonSchema":    "JSONSchema",
			"UuidValue":     "UUIDValue",
			"GuidValue":     "GUIDValue",
			"UidValue":      "UIDValue",
			"GidValue":      "GIDValue",
			"UrlValue":      "URLValue",
			"UriValue":      "URIValue",
			"UserId":        "UserID",
			"IpAddress":     "IPAddress",
			"TcpSocket":     "TCPSocket",
			"UdpSocket":     "UDPSocket",
			"TlsConfig":     "TLSConfig",
			"SshClient":     "SSHClient",
			"SqlQuery":      "SQLQuery",
			"RpcClient":     "RPCClient",
			"DnsResolver":   "DNSResolver",
			"DnssecRecord":  "DNSSECRecord",
			"JwtToken":      "JWTToken",
			"OidcProvider":  "OIDCProvider",
			"JwkKey":        "JWKKey",
			"JwksEndpoint":  "JWKSEndpoint",
			"SsoLogin":      "SSOLogin",
			"SamlResponse":  "SAMLResponse",
			"LdapClient":    "LDAPClient",
			"TomlConfig":    "TOMLConfig",
			"CsvReader":     "CSVReader",
			"TsvReader":     "TSVReader",
			"IniConfig":     "INIConfig",
			"MdRenderer":    "MDRenderer",
			"HtmlTemplate":  "HTMLTemplate",
			"CssTheme":      "CSSTheme",
			"JsBundle":      "JSBundle",
			"TsConfig":      "TSConfig",
			"OciImage":      "OCIImage",
			"SbomReport":    "SBOMReport",
			"K8sCluster":    "K8SCluster",
			"VcsRoot":       "VCSRoot",
			"ScmProvider":   "SCMProvider",
			"IacPlan":       "IaCPlan",
			"HmacKey":       "HMACKey",
			"AesCipher":     "AESCipher",
			"RsaKey":        "RSAKey",
			"EcdsaSigner":   "ECDSASigner",
			"PemBlock":      "PEMBlock",
			"DerBytes":      "DERBytes",
			"CsrRequest":    "CSRRequest",
			"CaBundle":      "CABundle",
			"KmsKey":        "KMSKey",
			"OsConfig":      "OSConfig",
			"CpuLimit":      "CPULimit",
			"RamUsage":      "RAMUsage",
			"PosixPath":     "POSIXPath",
			"AbiCompat":     "ABICompat",
			"PidFile":       "PIDFile",
			"EtlJob":        "ETLJob",
			"EltPipeline":   "ELTPipeline",
			"PkiBundle":     "PKIBundle",
			"AmqpClient":    "AMQPClient",
			"MqttBroker":    "MQTTBroker",
			"NatsSubject":   "NATSSubject",
			"RestClient":    "RESTClient",
			"SoapEnvelope":  "SOAPEnvelope",
			"TtySession":    "TTYSession",
			"EofReader":     "EOFReader",
		}

		for input, expected := range testCases {
			value, err := ApplyKnownAcronyms(input)
			require.NoError(t, err)
			assert.Equal(t, expected, value)
		}
	})

	t.Run("replaces configured acronym in camel case", func(t *testing.T) {
		value, err := ApplyKnownAcronyms("apiClient")
		require.NoError(t, err)
		assert.Equal(t, "apiClient", value)
	})

	t.Run("replaces consecutive configured acronyms in camel case", func(t *testing.T) {
		value, err := ApplyKnownAcronyms("openAiApiKey")
		require.NoError(t, err)
		assert.Equal(t, "openAIAPIKey", value)
	})

	t.Run("preserves identifier family words", func(t *testing.T) {
		value, err := ApplyKnownAcronyms("MaddieIdentifier")
		require.NoError(t, err)
		assert.Equal(t, "MaddieIdentifier", value)
	})

	t.Run("preserves idempotent family words", func(t *testing.T) {
		value, err := ApplyKnownAcronyms("idempotentRetry")
		require.NoError(t, err)
		assert.Equal(t, "idempotentRetry", value)
	})

	t.Run("still rewrites explicit id token", func(t *testing.T) {
		value, err := ApplyKnownAcronyms("adrienId")
		require.NoError(t, err)
		assert.Equal(t, "adrienID", value)
	})

	t.Run("preserves already uppercase acronym runs", func(t *testing.T) {
		value, err := ApplyKnownAcronyms("OpenAIAPIKey")
		require.NoError(t, err)
		assert.Equal(t, "OpenAIAPIKey", value)
	})

	t.Run("preserves adjacent uppercase acronym runs", func(t *testing.T) {
		value, err := ApplyKnownAcronyms("AIAPI")
		require.NoError(t, err)
		assert.Equal(t, "AIAPI", value)
	})

	t.Run("preserves already uppercase acronym runs in camel case", func(t *testing.T) {
		value, err := ApplyKnownAcronyms("openAIAPIKey")
		require.NoError(t, err)
		assert.Equal(t, "openAIAPIKey", value)
	})

	t.Run("preserves configured ai exception words", func(t *testing.T) {
		testCases := map[string]string{
			"acquaint":      "acquaint",
			"emailAddress":  "emailAddress",
			"maintainState": "maintainState",
			"railwayTrack":  "railwayTrack",
		}

		for input, expected := range testCases {
			value, err := ApplyKnownAcronyms(input)
			require.NoError(t, err)
			assert.Equal(t, expected, value)
		}
	})

	t.Run("preserves configured api exception words", func(t *testing.T) {
		testCases := map[string]string{
			"apiaryLocation":  "apiaryLocation",
			"apicultureGuide": "apicultureGuide",
			"capillaryFlow":   "capillaryFlow",
			"vapidityScore":   "vapidityScore",
		}

		for input, expected := range testCases {
			value, err := ApplyKnownAcronyms(input)
			require.NoError(t, err)
			assert.Equal(t, expected, value)
		}
	})

	t.Run("preserves configured hal exception words", func(t *testing.T) {
		testCases := map[string]string{
			"challengeMode":  "challengeMode",
			"halfTimeReport": "halfTimeReport",
			"inhaleRate":     "inhaleRate",
			"whalingLog":     "whalingLog",
		}

		for input, expected := range testCases {
			value, err := ApplyKnownAcronyms(input)
			require.NoError(t, err)
			assert.Equal(t, expected, value)
		}
	})

	t.Run("preserves configured cli exception words", func(t *testing.T) {
		testCases := map[string]string{
			"clientConfig":  "clientConfig",
			"climateData":   "climateData",
			"declineReason": "declineReason",
			"inclineAngle":  "inclineAngle",
		}

		for input, expected := range testCases {
			value, err := ApplyKnownAcronyms(input)
			require.NoError(t, err)
			assert.Equal(t, expected, value)
		}
	})

	t.Run("preserves configured rag exception words", func(t *testing.T) {
		testCases := map[string]string{
			"fragileState":    "fragileState",
			"fragranceNotes":  "fragranceNotes",
			"pragmaticChoice": "pragmaticChoice",
			"dragonScale":     "dragonScale",
		}

		for input, expected := range testCases {
			value, err := ApplyKnownAcronyms(input)
			require.NoError(t, err)
			assert.Equal(t, expected, value)
		}
	})

	t.Run("preserves configured sdk cli acronym proposal exception words", func(t *testing.T) {
		testCases := map[string]string{
			"luxuryTheme":      "luxuryTheme",
			"interestingValue": "interestingValue",
			"restaurantGuide":  "restaurantGuide",
			"initialConfig":    "initialConfig",
			"minimumValue":     "minimumValue",
			"markdownFile":     "markdownFile",
			"modernMode":       "modernMode",
			"cacheSize":        "cacheSize",
			"cancelReason":     "cancelReason",
			"cardiacMonitor":   "cardiacMonitor",
			"espressoShot":     "espressoShot",
			"soapstoneColor":   "soapstoneColor",
			"meltValue":        "meltValue",
		}

		for input, expected := range testCases {
			value, err := ApplyKnownAcronyms(input)
			require.NoError(t, err)
			assert.Equal(t, expected, value)
		}
	})

	t.Run("leaves unknown words alone apart from case normalisation", func(t *testing.T) {
		value, err := ApplyKnownAcronyms("sourceName")
		require.NoError(t, err)
		assert.Equal(t, "sourceName", value)
	})
}
