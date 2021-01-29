package api

import (
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"net/http"

	"github.com/julienschmidt/httprouter"

	"gitlab.com/NebulousLabs/Sia/build"
	"gitlab.com/NebulousLabs/Sia/crypto"
	"gitlab.com/NebulousLabs/Sia/modules"
	"gitlab.com/NebulousLabs/Sia/types"
	"gitlab.com/NebulousLabs/encoding"
)

// ConsensusGET contains general information about the consensus set, with tags
// to support idiomatic json encodings.
type ConsensusGET struct {
	// Consensus status values.
	Synced       bool              `json:"synced"`
	Height       types.BlockHeight `json:"height"`
	CurrentBlock types.BlockID     `json:"currentblock"`
	Target       types.Target      `json:"target"`
	Difficulty   types.Currency    `json:"difficulty"`

	// Foundation unlock hashes.
	FoundationPrimaryUnlockHash  types.UnlockHash `json:"foundationprimaryunlockhash"`
	FoundationFailsafeUnlockHash types.UnlockHash `json:"foundationfailsafeunlockhash"`

	// Consensus code constants.
	BlockFrequency         types.BlockHeight `json:"blockfrequency"`
	BlockSizeLimit         uint64            `json:"blocksizelimit"`
	ExtremeFutureThreshold types.Timestamp   `json:"extremefuturethreshold"`
	FutureThreshold        types.Timestamp   `json:"futurethreshold"`
	GenesisTimestamp       types.Timestamp   `json:"genesistimestamp"`
	MaturityDelay          types.BlockHeight `json:"maturitydelay"`
	MedianTimestampWindow  uint64            `json:"mediantimestampwindow"`
	SiafundCount           types.Currency    `json:"siafundcount"`
	SiafundPortion         *big.Rat          `json:"siafundportion"`

	InitialCoinbase uint64 `json:"initialcoinbase"`
	MinimumCoinbase uint64 `json:"minimumcoinbase"`

	RootTarget types.Target `json:"roottarget"`
	RootDepth  types.Target `json:"rootdepth"`

	SiacoinPrecision types.Currency `json:"siacoinprecision"`
}

type ConsensusBlocksGet struct {
	ID               types.BlockID                              `json:"id"`
	Height           types.BlockHeight                          `json:"height"`
	ParentID         types.BlockID                              `json:"parentid"`
	Nonce            types.BlockNonce                           `json:"nonce"`
	Timestamp        types.Timestamp                            `json:"timestamp"`
	MinerPayouts     []types.SiacoinOutput                      `json:"minerpayouts"`
	Transactions     []types.Transaction                        `json:"transactions"`
	TransactionIDs   []types.TransactionID                      `json:"transactionids"`
	SiacoinOutputIDs map[string]types.SiacoinOutputID `json:"siacoinoutputids"`
}

// ConsensusHeadersGET contains information from a blocks header.
type ConsensusHeadersGET struct {
	BlockID types.BlockID `json:"blockid"`
}

// ConsensusBlocksGet contains all fields of a types.Block and additional
// fields for ID and Height.
type ConsensusBlocksGet struct {
	ID           types.BlockID           `json:"id"`
	Height       types.BlockHeight       `json:"height"`
	ParentID     types.BlockID           `json:"parentid"`
	Nonce        types.BlockNonce        `json:"nonce"`
	Difficulty   types.Currency          `json:"difficulty"`
	Timestamp    types.Timestamp         `json:"timestamp"`
	MinerPayouts []types.SiacoinOutput   `json:"minerpayouts"`
	Transactions []ConsensusBlocksGetTxn `json:"transactions"`
}

// ConsensusBlocksGetTxn contains all fields of a types.Transaction and an
// additional ID field.
type ConsensusBlocksGetTxn struct {
	ID                    types.TransactionID               `json:"id"`
	SiacoinInputs         []types.SiacoinInput              `json:"siacoininputs"`
	SiacoinOutputs        []ConsensusBlocksGetSiacoinOutput `json:"siacoinoutputs"`
	FileContracts         []ConsensusBlocksGetFileContract  `json:"filecontracts"`
	FileContractRevisions []types.FileContractRevision      `json:"filecontractrevisions"`
	StorageProofs         []types.StorageProof              `json:"storageproofs"`
	SiafundInputs         []types.SiafundInput              `json:"siafundinputs"`
	SiafundOutputs        []ConsensusBlocksGetSiafundOutput `json:"siafundoutputs"`
	MinerFees             []types.Currency                  `json:"minerfees"`
	ArbitraryData         [][]byte                          `json:"arbitrarydata"`
	TransactionSignatures []types.TransactionSignature      `json:"transactionsignatures"`
}

// ConsensusBlocksGetFileContract contains all fields of a types.FileContract
// and an additional ID field.
type ConsensusBlocksGetFileContract struct {
	ID                 types.FileContractID              `json:"id"`
	FileSize           uint64                            `json:"filesize"`
	FileMerkleRoot     crypto.Hash                       `json:"filemerkleroot"`
	WindowStart        types.BlockHeight                 `json:"windowstart"`
	WindowEnd          types.BlockHeight                 `json:"windowend"`
	Payout             types.Currency                    `json:"payout"`
	ValidProofOutputs  []ConsensusBlocksGetSiacoinOutput `json:"validproofoutputs"`
	MissedProofOutputs []ConsensusBlocksGetSiacoinOutput `json:"missedproofoutputs"`
	UnlockHash         types.UnlockHash                  `json:"unlockhash"`
	RevisionNumber     uint64                            `json:"revisionnumber"`
}

// ConsensusBlocksGetSiacoinOutput contains all fields of a types.SiacoinOutput
// and an additional ID field.
type ConsensusBlocksGetSiacoinOutput struct {
	ID         types.SiacoinOutputID `json:"id"`
	Value      types.Currency        `json:"value"`
	UnlockHash types.UnlockHash      `json:"unlockhash"`
}

// ConsensusBlocksGetSiafundOutput contains all fields of a types.SiafundOutput
// and an additional ID field.
type ConsensusBlocksGetSiafundOutput struct {
	ID         types.SiafundOutputID `json:"id"`
	Value      types.Currency        `json:"value"`
	UnlockHash types.UnlockHash      `json:"unlockhash"`
}

// ConsensusBlocksGetFromBlock is a helper method that uses a types.Block, types.BlockHeight and
// types.Currency to create a ConsensusBlocksGet object.
func consensusBlocksGetFromBlock(b types.Block, h types.BlockHeight, d types.Currency) ConsensusBlocksGet {
	txns := make([]ConsensusBlocksGetTxn, 0, len(b.Transactions))
	for _, t := range b.Transactions {
		// Get the transaction's SiacoinOutputs.
		scos := make([]ConsensusBlocksGetSiacoinOutput, 0, len(t.SiacoinOutputs))
		for i, sco := range t.SiacoinOutputs {
			scos = append(scos, ConsensusBlocksGetSiacoinOutput{
				ID:         t.SiacoinOutputID(uint64(i)),
				Value:      sco.Value,
				UnlockHash: sco.UnlockHash,
			})
		}
		// Get the transaction's SiafundOutputs.
		sfos := make([]ConsensusBlocksGetSiafundOutput, 0, len(t.SiafundOutputs))
		for i, sfo := range t.SiafundOutputs {
			sfos = append(sfos, ConsensusBlocksGetSiafundOutput{
				ID:         t.SiafundOutputID(uint64(i)),
				Value:      sfo.Value,
				UnlockHash: sfo.UnlockHash,
			})
		}
		// Get the transaction's FileContracts.
		fcos := make([]ConsensusBlocksGetFileContract, 0, len(t.FileContracts))
		for i, fc := range t.FileContracts {
			// Get the FileContract's valid proof outputs.
			fcid := t.FileContractID(uint64(i))
			vpos := make([]ConsensusBlocksGetSiacoinOutput, 0, len(fc.ValidProofOutputs))
			for j, vpo := range fc.ValidProofOutputs {
				vpos = append(vpos, ConsensusBlocksGetSiacoinOutput{
					ID:         fcid.StorageProofOutputID(types.ProofValid, uint64(j)),
					Value:      vpo.Value,
					UnlockHash: vpo.UnlockHash,
				})
			}
			// Get the FileContract's missed proof outputs.
			mpos := make([]ConsensusBlocksGetSiacoinOutput, 0, len(fc.MissedProofOutputs))
			for j, mpo := range fc.MissedProofOutputs {
				mpos = append(mpos, ConsensusBlocksGetSiacoinOutput{
					ID:         fcid.StorageProofOutputID(types.ProofMissed, uint64(j)),
					Value:      mpo.Value,
					UnlockHash: mpo.UnlockHash,
				})
			}
			fcos = append(fcos, ConsensusBlocksGetFileContract{
				ID:                 fcid,
				FileSize:           fc.FileSize,
				FileMerkleRoot:     fc.FileMerkleRoot,
				WindowStart:        fc.WindowStart,
				WindowEnd:          fc.WindowEnd,
				Payout:             fc.Payout,
				ValidProofOutputs:  vpos,
				MissedProofOutputs: mpos,
				UnlockHash:         fc.UnlockHash,
				RevisionNumber:     fc.RevisionNumber,
			})
		}
		txns = append(txns, ConsensusBlocksGetTxn{
			ID:                    t.ID(),
			SiacoinInputs:         t.SiacoinInputs,
			SiacoinOutputs:        scos,
			FileContracts:         fcos,
			FileContractRevisions: t.FileContractRevisions,
			StorageProofs:         t.StorageProofs,
			SiafundInputs:         t.SiafundInputs,
			SiafundOutputs:        sfos,
			MinerFees:             t.MinerFees,
			ArbitraryData:         t.ArbitraryData,
			TransactionSignatures: t.TransactionSignatures,
		})
	}
	return ConsensusBlocksGet{
		ID:           b.ID(),
		Height:       h,
		ParentID:     b.ParentID,
		Nonce:        b.Nonce,
		Difficulty:   d,
		Timestamp:    b.Timestamp,
		MinerPayouts: b.MinerPayouts,
		Transactions: txns,
	}
}

// consensusHandler handles the API calls to /consensus.
func (api *API) consensusHandler(w http.ResponseWriter, _ *http.Request, _ httprouter.Params) {
	height := api.cs.Height()
	b, found := api.cs.BlockAtHeight(height)
	if !found {
		err := "Failed to fetch block for current height"
		WriteError(w, Error{err}, http.StatusInternalServerError)
		build.Critical(err)
		return
	}
	cbid := b.ID()
	currentTarget, _ := api.cs.ChildTarget(cbid)
	primary, failsafe := api.cs.FoundationUnlockHashes()
	WriteJSON(w, ConsensusGET{
		Synced:       api.cs.Synced(),
		Height:       height,
		CurrentBlock: cbid,
		Target:       currentTarget,
		Difficulty:   currentTarget.Difficulty(),

		FoundationPrimaryUnlockHash:  primary,
		FoundationFailsafeUnlockHash: failsafe,

		BlockFrequency:         types.BlockFrequency,
		BlockSizeLimit:         types.BlockSizeLimit,
		ExtremeFutureThreshold: types.ExtremeFutureThreshold,
		FutureThreshold:        types.FutureThreshold,
		GenesisTimestamp:       types.GenesisTimestamp,
		MaturityDelay:          types.MaturityDelay,
		MedianTimestampWindow:  types.MedianTimestampWindow,
		SiafundCount:           types.SiafundCount,
		SiafundPortion:         types.SiafundPortion,

		InitialCoinbase: types.InitialCoinbase,
		MinimumCoinbase: types.MinimumCoinbase,

		RootTarget: types.RootTarget,
		RootDepth:  types.RootDepth,

		SiacoinPrecision: types.SiacoinPrecision,
	})
}

// consensusBlocksIDHandler handles the API calls to /consensus/blocks
// endpoint.
func (api *API) consensusBlocksHandler(w http.ResponseWriter, req *http.Request, _ httprouter.Params) {
	// Get query params and check them.
	id, height := req.FormValue("id"), req.FormValue("height")
	if id != "" && height != "" {
		WriteError(w, Error{"can't specify both id and height"}, http.StatusBadRequest)
		return
	}
	if id == "" && height == "" {
		WriteError(w, Error{"either id or height has to be provided"}, http.StatusBadRequest)
		return
	}

	var b types.Block
	var h types.BlockHeight
	var exists bool
	var blockheight types.BlockHeight

	// Handle request by id
	if id != "" {
		var bid types.BlockID
		if err := bid.LoadString(id); err != nil {
			WriteError(w, Error{"failed to unmarshal blockid"}, http.StatusBadRequest)
			return
		}
		b, blockheight, exists = api.cs.BlockByID(bid)
	}
	// Handle request by height
	if height != "" {
		if _, err := fmt.Sscan(height, &h); err != nil {
			WriteError(w, Error{"failed to parse block height"}, http.StatusBadRequest)
			return
		}
		b, exists = api.cs.BlockAtHeight(types.BlockHeight(h))
		blockheight = types.BlockHeight(h)
	}
	// Check if block was found
	if !exists {
		WriteError(w, Error{"block doesn't exist"}, http.StatusBadRequest)
		return
	}

	var transactionIDs []types.TransactionID
	siacoinOutputIDs := make(map[string]types.SiacoinOutputID)

	for _, txn := range b.Transactions {
		txid := txn.ID()
		transactionIDs = append(transactionIDs, txid)
		for j := range txn.SiacoinOutputs {
			key := fmt.Sprintf("%s_%d",txid,j)
			siacoinOutputIDs[key] = txn.SiacoinOutputID(uint64(j))
		}
	}



	//for i,txn := range b.MinerPayouts {
	//	unlockHash = b.MinerPayouts[i].UnlockHash
	//	outputID = b.MinerPayouts[i] ?
	//}

	// Write response
	WriteJSON(w, ConsensusBlocksGet{
		b.ID(),
		blockheight,
		b.ParentID,
		b.Nonce,
		b.Timestamp,
		b.MinerPayouts,
		b.Transactions,
		transactionIDs,
		siacoinOutputIDs,
	})
}

// consensusValidateTransactionsetHandler handles the API calls to
// /consensus/validate/transactionset.
func (api *API) consensusValidateTransactionsetHandler(w http.ResponseWriter, req *http.Request, _ httprouter.Params) {
	var txnset []types.Transaction
	err := json.NewDecoder(req.Body).Decode(&txnset)
	if err != nil {
		WriteError(w, Error{"could not decode transaction set: " + err.Error()}, http.StatusBadRequest)
		return
	}
	_, err = api.cs.TryTransactionSet(txnset)
	if err != nil {
		WriteError(w, Error{"transaction set validation failed: " + err.Error()}, http.StatusBadRequest)
		return
	}
	WriteSuccess(w)
}

// consensusSubscribeHandler handles the API calls to the /consensus/subscribe
// endpoint.
func (api *API) consensusSubscribeHandler(w http.ResponseWriter, req *http.Request, ps httprouter.Params) {
	var ccid modules.ConsensusChangeID
	if err := (*crypto.Hash)(&ccid).LoadString(ps.ByName("id")); err != nil {
		WriteError(w, Error{"could not decode ID: " + err.Error()}, http.StatusBadRequest)
		return
	}

	// create subscriber and start processing changes in a goroutine
	errCh := make(chan error, 1)
	ccs := newConsensusChangeStreamer(w)
	go func() {
		errCh <- api.cs.ConsensusSetSubscribe(ccs, ccid, req.Context().Done())
		api.cs.Unsubscribe(ccs)
	}()
	err := <-errCh
	if err != nil {
		// TODO: we can't call WriteError here; the client is expecting binary.
		return
	}
}

type consensusChangeStreamer struct {
	e *encoding.Encoder
}

func (ccs consensusChangeStreamer) ProcessConsensusChange(cc modules.ConsensusChange) {
	ccs.e.Encode(cc)
}

func newConsensusChangeStreamer(w io.Writer) consensusChangeStreamer {
	return consensusChangeStreamer{
		e: encoding.NewEncoder(w),
	}
}
