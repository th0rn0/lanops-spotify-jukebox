package main

func getNextSongByVotes() (Track, error) {
	var track Track

	// if err := db.Raw("SELECT MAX(votes) FROM tracks").First(&track).Error; err != nil {
	if err := db.Raw("SELECT * FROM tracks WHERE votes = ( SELECT MAX(votes) FROM tracks )").First(&track).Error; err != nil {
		return track, err
	}
	return track, nil
}
