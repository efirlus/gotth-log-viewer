// ==UserScript==
// @name         Auto Download with Ctrl+Shift+D (Download + Rename)
// @namespace    http://tampermonkey.net/
// @version      1.3
// @description  Click to reveal download link, auto-download, and rename file based on page content (Ctrl+Shift+D)
// @author       Your Name
// @match        *://*/*
// @grant        none
// ==/UserScript==


// ==UserScript==
// @name         Auto Download Downlink (with GM_download)
// @namespace    http://tampermonkey.net/
// @version      0.3
// @description  Automatically download file from specific download button when Ctrl+Shift+D is pressed, using GM_download
// @match        *://*/*
// @grant        GM_download
// ==/UserScript==

(function() {
    'use strict';

    // Listen for the keyboard shortcut: Ctrl+Shift+D
    document.addEventListener('keydown', function(e) {
        // Check for Ctrl + Shift + D
        if (e.ctrlKey && e.shiftKey && e.code === 'KeyD') {
            e.preventDefault(); // Prevent default browser behavior

            // Step 2: Find and click the <a class="pa-2 download-btn"> button to reveal the download link
            const downloadButton = document.querySelector('a.pa-2.download-btn');
            if (downloadButton) {
                downloadButton.click();

                // Step 3: Wait for the download link to be revealed, then find the correct link containing "SOURCE"
                setTimeout(function() {
                    // Find the <a href> that contains <span class="v-btn__content"> with "SOURCE"
                    const sourceLink = Array.from(document.querySelectorAll('a[href] > button > span.v-btn__content'))
                                           .find(span => span.textContent.includes('SOURCE'));

                    if (sourceLink && sourceLink.closest('a')) {
                        const href = sourceLink.closest('a').getAttribute('href');
                        if (href) {
                            // Step 4: Get the <h1 class="pl-2"> element's text for renaming the file
                            const titleElement = document.querySelector('h1.pl-2');
                            let fileName = 'downloaded_file'; // Default file name
                            if (titleElement && titleElement.textContent) {
                                // Clean up the title string and use it as the file name
                                fileName = titleElement.textContent.trim().replace(/[/\\?%*:|"<>]/g, '_');
                            }

                            // Step 5: Use GM_download to trigger the download
                            GM_download({
                                url: href,
                                name: fileName,
                                saveAs: true // This will open the "Save As" dialog
                            });
                        } else {
                            console.error('No href attribute found.');
                        }
                    } else {
                        console.error('Download link with "SOURCE" not found.');
                    }
                }, 1000); // Adjust the delay if needed to wait for the download link to appear
            } else {
                console.error('Download button not found.');
            }
        }
    });
})();