package event

func TestNew(t *testing.T) {
	d := faker.UUID()
	e := New(d)
	assert.Equal(t, d, e.Data())
}