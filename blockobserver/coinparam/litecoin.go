package coinparam

import (
	"time"

	"github.com/mit-dci/lit/btcutil/chaincfg/chainhash"
	"github.com/mit-dci/pooldetective/blockobserver/wire"

	"golang.org/x/crypto/scrypt"
)

// LiteCoinTestNet4Params are the parameters for the litecoin test network 4.
var LitecoinParams = Params{
	Name:          "litecoin",
	Ticker:        "LTC",
	CoinID:        7,
	NetMagicBytes: 0xdbb6c0fb,
	DefaultPort:   "9333",
	DNSSeeds: []string{
		"dnsseed.litecointools.com",
		"dnsseed.litecoinpool.org",
		"dnsseed.koin-project.com",
		"seed-a.litecoin.loshan.co.uk",
		"dnsseed.thrasher.io",
	},

	// Chain parameters
	GenesisBlock: &liteCoinGenesisBlock,
	GenesisHash:  &liteCoinGenesisHash,

	PoWFunction: func(b []byte, height int32) chainhash.Hash {
		scryptBytes, _ := scrypt.Key(b, b, 1024, 1, 1, 32)
		asChainHash, _ := chainhash.NewHash(scryptBytes)
		return *asChainHash
	},
	DiffCalcFunction: diffBitcoin,
	StartHeader: [80]byte{
		0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0xd9, 0xce, 0xd4, 0xed, 0x11, 0x30, 0xf7, 0xb7, 0xfa, 0xad, 0x9b, 0xe2,
		0x53, 0x23, 0xff, 0xaf, 0xa3, 0x32, 0x32, 0xa1, 0x7c, 0x3e, 0xdf, 0x6c,
		0xfd, 0x97, 0xbe, 0xe6, 0xba, 0xfb, 0xdd, 0x97, 0xf6, 0x0b, 0xa1, 0x58,
		0xf0, 0xff, 0x0f, 0x1e, 0xe1, 0x79, 0x04, 0x00,
	},
	StartHeight:              48384,
	AssumeDiffBefore:         50401,
	FeePerByte:               800,
	PowLimit:                 liteCoinTestNet4PowLimit,
	PowLimitBits:             0x1e0fffff,
	CoinbaseMaturity:         100,
	SubsidyReductionInterval: 840000,
	TargetTimespan:           time.Hour * 84,    // 84 hours
	TargetTimePerBlock:       time.Second * 150, // 150 seconds
	RetargetAdjustmentFactor: 4,                 // 25% less, 400% more
	ReduceMinDifficulty:      true,
	MinDiffReductionTime:     time.Minute * 10, // ?? unknown
	GenerateSupported:        false,

	// Checkpoints ordered from oldest to newest.
	Checkpoints: []Checkpoint{},

	// Enforce current block version once majority of the network has
	// upgraded.
	// 51% (51 / 100)
	// Reject previous block versions once a majority of the network has
	// upgraded.
	// 75% (75 / 100)
	BlockEnforceNumRequired: 51,
	BlockRejectNumRequired:  75,
	BlockUpgradeNumToCheck:  100,

	// Mempool parameters
	RelayNonStdTxs: true,

	// Address encoding magics
	PubKeyHashAddrID: 0x6f, // starts with m or n
	ScriptHashAddrID: 0xc4, // starts with 2
	Bech32Prefix:     "tltc",
	PrivateKeyID:     0xef, // starts with 9 7(uncompressed) or c (compressed)

	// BIP32 hierarchical deterministic extended key magics
	HDPrivateKeyID: [4]byte{0x04, 0x35, 0x83, 0x94}, // starts with tprv
	HDPublicKeyID:  [4]byte{0x04, 0x35, 0x87, 0xcf}, // starts with tpub

	// BIP44 coin type used in the hierarchical deterministic path for
	// address generation.
	HDCoinType:        65538, // i dunno, 0x010001 ?
	IdentifyAsClient:  "LitecoinCashCore",
	IdentifyAsVersion: "0.17.1",
}

// liteCoinTestNet4GenesisBlock has is like completely its own thing
var liteCoinGenesisBlock = wire.MsgBlock{
	Header: wire.BlockHeader{
		Version:    1,
		PrevBlock:  chainhash.Hash{}, // empty
		MerkleRoot: liteCoinMerkleRoot,
		Timestamp:  time.Unix(1231006505, 0), // later
		Bits:       0x1d00ffff,
		Nonce:      2083236893,
	},
	//	Transactions: []*wire.MsgTx{&genesisCoinbaseTx}, // this is wrong... will it break?
}

var liteCoinGenesisHash = chainhash.Hash([chainhash.HashSize]byte{
	0xe2, 0xbf, 0x04, 0x7e, 0x7e, 0x5a, 0x19, 0x1a, 0xa4, 0xef, 0x34, 0xd3, 0x14, 0x97, 0x9d, 0xc9, 0x98, 0x6e, 0x0f, 0x19, 0x25, 0x1e, 0xda, 0xba, 0x59, 0x40, 0xfd, 0x1f, 0xe3, 0x65, 0xa7, 0x12,
})

var liteCoinMerkleRoot = chainhash.Hash([chainhash.HashSize]byte{
	0xd9, 0xce, 0xd4, 0xed, 0x11, 0x30, 0xf7, 0xb7, 0xfa, 0xad, 0x9b, 0xe2, 0x53, 0x23, 0xff, 0xaf, 0xa3, 0x32, 0x32, 0xa1, 0x7c, 0x3e, 0xdf, 0x6c, 0xfd, 0x97, 0xbe, 0xe6, 0xba, 0xfb, 0xdd, 0x97,
})
