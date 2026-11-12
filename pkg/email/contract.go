
package email

import "context"

type Message struct {
    From    string
    To      []string
    Cc      []string
    Bcc     []string
    ReplyTo string
    Subject string
    HTML    string
    Text    string
}

type Sender interface {
    Send(ctx context.Context, message Message) error
}
