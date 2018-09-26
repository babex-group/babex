package babex

type Meta map[string]string

func (m Meta) Merge(metas ...Meta) {
	for _, meta := range metas {
		for key, item := range meta {
			m[key] = item
		}
	}
}
