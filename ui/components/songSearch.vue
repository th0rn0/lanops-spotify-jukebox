<template>
    <form @submit.prevent="searchTracks">
        <div class="input-group mb-3">
            <input v-model="search" class="form-control" placeholder="Search">
            <button class="btn btn-primary" type="submit">Search</button>
        </div>
    </form>
    <table class="table table-striped">
        <tbody>
            <SearchItem v-for="track in searchResult.track" :track="track" />
        </tbody>
    </table>
</template>

<script lang="ts" setup>

    const search = ref("")
    const searchResult = ref({});

    async function searchTracks() {
        searchResult.value = await $fetch(`http://localhost:8888/search/${search.value}`).catch((error) => error.data)
    }
</script>