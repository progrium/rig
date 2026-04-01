package node

// Destroy walks the subtree rooted at n (depth-first, exit pass only). For each node that
// implements RealmNode, it removes that node from its parent's subnode list when a parent
// exists, then calls Realm.Destroy on the node's realm. Nodes without a realm are skipped.
// It returns any error from Walk, subnode removal, or Destroy.
func Destroy(n Node) error {
	return Walk(n, nil, func(nn Node) error {
		if rn, ok := Unwrap[RealmNode](nn); ok {
			p := Parent(nn)
			if p != nil {
				if err := RemoveSubnodeID(p, Kind(nn), nn.NodeID()); err != nil {
					return err
				}
			}
			return rn.GetRealm().Destroy(nn)
		}
		return nil
	})
}
