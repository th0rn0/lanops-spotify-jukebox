package jukebox

func (c *Client) VoteToSkip() {
	c.skip.votes++
}

func (c *Client) shouldSkip() bool {
	if c.skip.active == true || c.skip.votes >= c.cfg.VoteCountToSkip {
		return true
	}
	return false
}

func (c *Client) resetSkip() {
	c.skip.active = false
	c.skip.votes = 0
}
