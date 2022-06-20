package histogram

import (
	patterns "bit-wordy/src/primitives"
	"fmt"
)

const Scale = 40

type Bar []patterns.Result

func (b Bar) String() string {
	render := fmt.Sprintf("count: %4d =| ", len(b))
	samples := make(Bar, len(b)/Scale)
	for i := range samples {
		if i*Scale < len(b) {
			samples[i] = b[i*Scale]
		}
	}

	for _, s := range samples {
		render += s.String() + "-"
	}

	return render + "\n"
}

type Histogram map[patterns.Pattern]Bar

func (h Histogram) String() string {
	s := ""
	for i := 0; i < (3 * 3 * 3 * 3 * 3); i++ {
		p := patterns.FromInt(i)
		if bar, exists := h[p]; exists {
			s += fmt.Sprintf("id: %4d - %s", p.Sum(), bar)
		}
	}

	return s
}
