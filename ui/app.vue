<template>
    <div class="container">
        <div class="row">
            <div class="col">
                <h3>Playlist</h3>
                <div v-for="track in playlist">
                    <img :src="track.images[2].url" />
                    {{ track.name }} - {{ track.artist }}
                </div>
            </div>
            <div class="col">
                <h3>Song Search</h3>
                <input v-model="search" class="form-control" id="exampleFormControlInput1" placeholder="">
                <button class="btn btn-primary" @click="searchTracks">Search</button>
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
