import { writable } from 'svelte/store'

/** The "Program arguments" text typed by the user; Run and Debug both use it (split like a shell). */
export const programArguments = writable('')
