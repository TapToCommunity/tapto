package state

type TokenQueue struct {
	Tokens chan Token
}

func (q *TokenQueue) Enqueue(t Token) {
	q.Tokens <- t
}

func (q *TokenQueue) Dequeue() Token {
	return <-q.Tokens
}

func (q *TokenQueue) Close() {
	close(q.Tokens)
}

func NewTokenQueue() *TokenQueue {
	return &TokenQueue{
		Tokens: make(chan Token),
	}
}
