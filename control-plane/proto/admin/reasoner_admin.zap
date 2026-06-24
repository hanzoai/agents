package admin

struct Reasoner {
    ReasonerId    text @0
    AgentNodeId   text @8
    Name          text @16
    Description   text @24
    Status        text @32
    NodeVersion   text @40
    LastHeartbeat text @48
}

struct ListReasonersResponse {
    Reasoners list<Reasoner> @0
}

interface AdminReasoner {
    listReasoners() returns (resp: ListReasonersResponse)
}
