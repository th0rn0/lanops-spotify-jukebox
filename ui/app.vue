<template>
    <div class="container">
        <div class="row">
            <div class="col">
                <h3>Playlist</h3>
                <div v-for="track in playlist">
                    <img :src="track.images[2].url" />
                    {{ track.name }} - {{ track.artist }}
                    <div class="btn-group">
                        <button class="btn btn-success">Vote up</button>
                        <button class="btn btn-danger">Vote down</button>
                    </div>
                </div>
            </div>
            <div class="col">
                <h3>Song Search</h3>
                <div class="input-group mb-3">
                    <input v-model="search" class="form-control" placeholder="Search">
                    <button class="btn btn-primary" type="button" @click="searchTracks">Search</button>
                </div>
                <div class="accordion" id="searchList">
                    <SearchItem v-for="track in searchResult.track" :track="track" />
                </div>
            </div>
        </div>
    </div>
</template>

<script lang="ts" setup>

    const search = ref("")
    const searchResult = ref({});
    const playlist = ref([]);

    onMounted(async () => {
        playlist.value = await $fetch(`http://localhost:8888/tracks`).catch((error) => error.data)
    });

    async function searchTracks() {
        searchResult.value = await $fetch(`http://localhost:8888/search/${search.value}`).catch((error) => error.data)
    }
</script>
