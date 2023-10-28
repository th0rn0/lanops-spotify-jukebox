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
        <td>
            <div class="btn-group-vertical" role="group" aria-label="Basic example">
                <button type="button" class="btn btn-success btn-sm"  @click="voteUp">Vote up</button>
                <button type="button" class="btn btn-danger btn-sm"  @click="voteDown">Vote down</button>
                </div>
        </td>
    </tr>
</template>

<script lang="ts" setup>
    const emit = defineEmits(["voted"]);
    
    export interface Props {
        track: {
            uri: string;
            name: string;
            artist: string;
            images: {
                height: number;
                width: number;
                url: string;
            }[]
        }
    }

    const props = defineProps<Props>()
    const runtimeConfig = useRuntimeConfig()

    async function voteUp() {
        await $fetch(`${runtimeConfig.public.apiEndpoint}/votes/add`, {method: 'POST', body: { uri: props.track.uri }})
        emit('voted');
    }

    async function voteDown() {
        await $fetch(`${runtimeConfig.public.apiEndpoint}/votes/remove`, {method: 'POST', body: { uri: props.track.uri }})
        emit('voted');
    }
</script>