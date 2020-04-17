package contractchain

/*
1. chainID = GetChainRandomlyFromDB()

2. voting = JudgeSafetyContractResultStateRootMinersFromSameMiners(chainID) {
		if mySafetyContractResultStateRootMiners[`ChainID`] == myPreviousSafetyContractResultStateRootMiners[`ChainID`] {
			return true
		}
		return false
	}

	if !voting {
		goto (7)
	}

3. popRelay = false, timeStamp = GetTime(), H = hash (timeStamp / RelaySwitchTimeUnit + chainID), relayList = SortRelayWithHash(H)
   Relay = GetFirstRelay(popRelay)
   peerID = SelectPeerRandomly()

   futureStateRoot = HamtGraphRelaySync( Relay, peerID, chainID, null, selector(field:=contractJSON));
	if err
		popRealy = true
		go to (5)

   myRelays[successed].add{this Relay}

4. stateRoots <-- TraverseHistoryStateRoot(futureStateRoot, one week time / block interval) {
		traverse history from futureStateRoot for states roots collection until one week time / block interval.
		(*)
		stateroot= y/contractJSON/SafetyContractResultStateRoot // recursive getting previous stateRoot to move into history
		y = HamtGraphRelaySync(stateroot)
		goto (*) until the mutable range or any error; //
	}

5. Relay = GetFirstRelay(popRelay)
   peerID = SelectPeerRandomly()
goto step (4) until surveyed 2/3 of myPeers[`ChainID`][...]

6. votingSafetyStateRoot = Voting(stateRoots) {
	list = CountAndSort(); // from high to low, low first when same count
	return = SelectFirstOneFromList();
}
update n+1 contract

7. graphRelaySync( Relay, peerID_A, chainID, null, selector(field:=contractJSON));
   if err= null, myRelays[successed].add{this Relay}

8. if ok, then verify {
if received ContractResultStateRoot/contractJSON shows a more difficult chain than SafetyContractResultStateRoot/contractJSON/`difficulty` and future root/contract/json/ safetyroot timestamp is passed clock, 不能在未来再次预测未来, then verify this chain's transactions until the MutableRange. in the verify process, it needs to add all db variables, hamt and amt trie to local. for some Key value, it will need `graphRelaySync` to get data from peerID_A;
goto (9)
}
  else {
  failed or err , if the (current time - safety state time ) is bigger than SelfMiningTime, then generate a new state on own  safety root, this will cause safety miner = previous safety miner to trigger voting, go to (9)
  };
       else go to step 1.

9. VerifyAndProcessContract() {
		Verify()
		ProcessContract()
	}
   update n+1 contract

10. goto (1)

*/
