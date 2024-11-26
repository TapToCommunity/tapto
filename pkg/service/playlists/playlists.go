package playlists

type Playlist struct {
	Media []string
	Index int
}

func NewPlaylist(media []string) *Playlist {
	return &Playlist{
		Media: media,
		Index: 0,
	}
}

func (p *Playlist) Next() string {
	p.Index++
	if p.Index >= len(p.Media) {
		p.Index = 0
	}
	return p.Media[p.Index]
}

func (p *Playlist) Previous() string {
	p.Index--
	if p.Index < 0 {
		p.Index = len(p.Media) - 1
	}
	return p.Media[p.Index]
}

func (p *Playlist) Current() string {
	return p.Media[p.Index]
}

type PlaylistController struct {
	Active *Playlist
	Queue  chan<- *Playlist
}
