package contractchain

/*
1. chainID = GetChainRandomlyFromDB() {
		dbChains->map[ChainID] config
	}
	BestBlock = GetBestBlockFromDB(chainID) {
		BestBlock->map[ChainID] Block
	}
	VotesCountingPoints = GetVotesCountingPointsBlock(chainID) {
		dbVotesCountingPoints->map[ChainID] root
	}

2. forging = IsTimeForging() {
		return ((time.Now() - BestBlock.TimeStamp) > MaxBlockTime) || POT()
	}
	if !forging {
		goto (7)
	}

3. timeStamp = GetTime(), H = hash (timeStamp / RelaySwitchTimeUnit + chainID)
	relay = ChooseRelay(H) {
		1:1:8
	}
	peerID = SelectPeerRandomly()

4. futureBlock = GraphSyncNPlusOneBlock()

5. if futureBlock.Difficulty > localBlock.Difficulty {
	Blocks = GraphSyncBlocks()
	VerifyBlock(block) {
		IfNPlusOne() || POT()
	}
	ProcessBlock(block) {
		// --------key------------value--------
		// 1-----address----------height-------
		// 2-----address_height--(state:preHeight, balance, nonce)-----
		// minerAddress
		lastChangeHeight = db.get(minerAddress)
		lastState = db.get(minerAddress_lastChangeHeight)
		newState.balance += lastState.Balance
		newState.preHeight = lastChangeHeight
		db.put(minerAddress_BlockHeight, newState)
		db.put(minerAddress, BlockHeight)
		// senderAddress
		lastChangeHeight = db.get(senderAddress)
		lastState = db.get(senderAddress_lastChangeHeight)
		newState.balance -= lastState.Balance
		newState.preHeight = lastChangeHeight
		newState.nonce += lastState.nonce
		db.put(senderAddress_BlockHeight, newState)
		db.put(senderAddress, BlockHeight)
		// receiverAddress
		lastChangeHeight = db.get(receiverAddress)
		lastState = db.get(receiverAddress_lastChangeHeight)
		newState.balance += lastState.Balance
		newState.preHeight = lastChangeHeight
		db.put(receiverAddress_BlockHeight, newState)
		db.put(receiverAddress, BlockHeight)
	}
	BlockRoots.add(Blocks.root)
}

6.	root = GraphSyncRandomHeight(height)
	BlockRoots.add(root)

7.	GenerateNPlusOneBlock(BestBlock) {
		for() {
			if BestBlock.TimeStamp < Time.Now()
				GenerateNextBlock(BestBlock) {
				}
		}
	}

8. voting = (BestBlock.BlockNumber - VotesCountingPoints.BlockNumber) > MutableRange

	if voting {
		VotesCountingPoints = Voting(BlockRoots) {
			list = CountAndSort(); // from high to low, low first when same count
			return = SelectFirstOneFromList();
		}
		updateVotesCountingPoints(VotesCountingPoints)
	}
	goto 1


*/
