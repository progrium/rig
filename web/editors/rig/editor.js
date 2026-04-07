
export const Editor = {
    view: ({attrs}) => {
        const {m,vscode,peer,realm,nodeID} = attrs;
        return m("div", "I am THE editor for node: " + nodeID);
    }
}
