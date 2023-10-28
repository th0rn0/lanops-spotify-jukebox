<template>
    <tr class="d-flex align-items-stretch">
        <td>
            <img :src="track.images[2].url" />
        </td>
        <td class="flex-grow-1 align-middle">
            <h5>
                {{ track.name }}
                <br />
                <small class="text-white-50">{{ track.artist }}</small>
            </h5>
        </td>
        <td class="align-middle">
            <button class="btn btn-primary" @click="addToPlaylist" >Add to playlist</button>
        </td>
    </tr>
</template>

<script lang="ts" setup>
    export interface Props {
        track: {
            id: string;
            name: string;
            artist: string;
            images: {
                height: number;
                width: number;
                url: string;
            }[]
        }
    }

    const emit = defineEmits(["songAdded"]);

    const props = defineProps<Props>()
    const runtimeConfig = useRuntimeConfig()

    async function addToPlaylist() {
        await $fetch(runtimeConfig.public.apiEndpoint + "/tracks/add", {method: 'POST', body: { uri: `spotify:track:${props.track.id}` }});
        emit('songAdded');
    }
</script>
