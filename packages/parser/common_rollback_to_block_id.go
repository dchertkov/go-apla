// Copyright 2016 The go-daylight Authors
// This file is part of the go-daylight library.
//
// The go-daylight library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-daylight library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-daylight library. If not, see <http://www.gnu.org/licenses/>.

package parser

import (
	"bytes"
	"database/sql"
	"strconv"

	"github.com/AplaProject/go-apla/packages/consts"
	"github.com/AplaProject/go-apla/packages/converter"
	"github.com/AplaProject/go-apla/packages/model"

	log "github.com/sirupsen/logrus"
)

// RollbackToBlockID rollbacks blocks till blockID
func (p *Parser) RollbackToBlockID(blockID int64) error {
	logger := p.GetLogger()
	_, err := model.MarkVerifiedAndNotUsedTransactionsUnverified()
	if err != nil {
		logger.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("marking verified and not used transactions unverified")
		return p.ErrInfo(err)
	}

	limit := 1000
	// roll back our blocks
	for {
		block := &model.Block{}
		blocks, err := block.GetBlocks(blockID, int32(limit))
		if err != nil {
			logger.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("getting blocks")
			return p.ErrInfo(err)
		}
		if len(blocks) == 0 {
			break
		}
		for _, block := range blocks {
			// roll back our blocks to the block blockID
			err = BlockRollback(block.Data)
			if err != nil {
				return p.ErrInfo(err)
			}
		}
		blocks = blocks[:0]
	}
	block := &model.Block{}
	_, err = block.Get(blockID)
	if err != nil && err != sql.ErrNoRows {
		logger.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("getting block")
		return p.ErrInfo(err)
	}

	isFirstBlock := blockID == 1
	header, err := ParseBlockHeader(bytes.NewBuffer(block.Data), !isFirstBlock)
	if err != nil {
		return p.ErrInfo(err)
	}

	ib := &model.InfoBlock{
		Hash:           block.Hash,
		BlockID:        header.BlockID,
		Time:           header.Time,
		EcosystemID:    header.EcosystemID,
		KeyID:          header.KeyID,
		NodePosition:   converter.Int64ToStr(header.NodePosition),
		CurrentVersion: strconv.Itoa(header.Version),
	}

	err = ib.Update(p.DbTransaction)
	if err != nil {
		logger.WithFields(log.Fields{"type": consts.DBError, "error": err}).Error("updating info block")
		return p.ErrInfo(err)
	}

	return nil
}
