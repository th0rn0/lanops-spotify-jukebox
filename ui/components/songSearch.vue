<template>
    <form @submit.prevent="searchTracks">
        <div class="input-group mb-3">
            <input v-model="search" class="form-control" placeholder="Search">
            <button class="btn btn-primary" type="submit">Search</button>
        </div>
    </form>
    <table class="table table-dark table-striped">
        <tbody>
            <SearchItem v-for="track in searchResult.track" :track="track" @songAdded="$emit('songAdded');" />
        </tbody>
    </table>
</template>

<script lang="ts" setup>

    const emit = defineEmits(["songAdded"]);

    const search = ref("")
    const searchResult = ref({});
    const runtimeConfig = useRuntimeConfig()

    async function searchTracks() {
        searchResult.value = await $fetch(runtimeConfig.public.apiEndpoint + `/search/${search.value}`).catch((error) => error.data)
    }
</script>