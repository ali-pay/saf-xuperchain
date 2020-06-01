package xchaincore

import (
	"fmt"
	"github.com/xuperchain/xuperchain/core/consensus/tdpos"
	"github.com/xuperchain/xuperchain/core/pb"
	"math/big"
	"strconv"
)

//生成区块的投票奖励
func (xc *XChainCore) GenerateVoteAward() ([]*pb.Transaction, error) {

	//奖励交易
	txs := make([]*pb.Transaction, 0)

	//选民的票数统计
	ballots := make(map[string]int64)

	//候选人列表
	candidates, err := xc.GetDposCandidates()
	if err != nil {
		xc.log.Error("[Vote_Award] fail to get candidates", "err", err)
		return nil, err
	}

	//遍历所有候选人
	for _, candidate := range candidates {
		records, err := xc.GetDposVotedRecords(candidate)
		if err != nil {
			xc.log.Error("[Vote_Award] fail to get voted records", "err", err)
			return nil, err
		}
		//xc.log.Warn("[Vote_Award] get voted records success", "current candidate is", candidate, "voted records", records)

		//遍历所有候选人的被投票记录，查询每笔投票的票数
		for _, voted := range records {
			//构造查询key
			keyVoteCandidate := tdpos.GenVoteCandidateKey(voted.Voter, candidate, voted.Txid)

			//查询投票票数
			balVal, err := xc.Utxovm.GetFromTable(nil, []byte(keyVoteCandidate))
			if err != nil {
				xc.log.Error("[Vote_Award] fail to get from table", "err", err)
				return nil, err
			}

			//票数
			ballot, err := strconv.ParseInt(string(balVal), 10, 64)
			if err != nil {
				xc.log.Error("[Vote_Award] fail to balVal parse int64", "err", err)
				return nil, err
			}

			//xc.log.Warn("[Vote_Award] get ballot success", "voter", voted.Voter, "ballot", ballot)

			ballots[voted.Voter] += ballot //统计选民的投票
			ballots["all"] += ballot       //统计全网总票数
		}
	}

	//打印票数
	for voter, ballot := range ballots {
		if voter == "all" {
			xc.log.Info("[Vote_Award] all ballots count", "ballots", ballots["all"])
			continue
		}
		xc.log.Info("[Vote_Award] voter ballots count", "voter", voter, "ballot", ballot)
	}

	//生成奖励
	for voter, ballot := range ballots {
		if voter == "all" {
			continue
		}

		//投票占比
		r := new(big.Rat)
		r.SetString(fmt.Sprintf("%d/%d", ballot, ballots["all"]))
		ratio, err := strconv.ParseFloat(r.FloatString(16), 10)
		if err != nil {
			xc.log.Error("[Vote_Award] fail to ratio parse float64", "err", err)
			return nil, err
		}

		//投票奖励
		voteAward := xc.Ledger.GenesisBlock.CalcVoteAward(tdpos.VoteAward, ratio)
		ratioStr := fmt.Sprintf("%.16f", ratio)
		xc.log.Info("[Vote_Award] calc vote award success", "voter", voter, "ratio", ratioStr, "award", voteAward)

		//奖励为0的不生成交易
		if voteAward.Int64() == 0 {
			continue
		}

		//生成交易
		voteawardtx, err := xc.Utxovm.GenerateVoteAwardTx([]byte(voter), voteAward.String(), []byte{'1'})
		if err != nil {
			xc.log.Error("[Vote_Award] fail to generate vote award tx", "err", err)
			return nil, err
		}
		txs = append(txs, voteawardtx)
	}

	return txs, nil
}
