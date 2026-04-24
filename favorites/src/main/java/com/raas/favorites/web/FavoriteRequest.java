package main.java.com.raas.favorites.web;

import jakarta.validation.constraints.NotBlank;

public class FavoriteRequest {
    @NotBlank
    private String userId;

    @NotBlank
    private String listingId;

    public String getUserId() {
        return userId;
    }

    public void setUserId(String userId) {
        this.userId = userId;
    }

    public String getListingId() {
        return listingId;
    }

    public void setListingId(String listingId) {
        this.listingId = listingId;
    }
}
