<template>
    <div class="container">
        <div class="row">
            <div class="col">
                <h3>Playlist</h3>
                <PlaylistItem />
            </div>
            <div class="col">
                <h3>Song Search</h3>
                <input v-model="search" class="form-control" id="exampleFormControlInput1" placeholder="">
                <button class="btn btn-primary" @click="searchTracks">Search</button>
            </div>
        </div>
    </div>
</template>

<script lang="ts" setup>

    const search = ref("")

    async function searchTracks() {
        const tracks = await $fetch(`http://localhost:8888/search/${search.value}`).catch((error) => error.data)
        console.log(tracks)
    }
</script>
