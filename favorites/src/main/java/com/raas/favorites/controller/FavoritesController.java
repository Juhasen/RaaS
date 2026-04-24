package main.java.com.raas.favorites.controller;

import com.raas.favorites.service.FavoritesService;
import com.raas.favorites.web.FavoriteRequest;
import jakarta.validation.Valid;
import java.util.List;
import org.springframework.http.HttpStatus;
import org.springframework.web.bind.annotation.DeleteMapping;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.PathVariable;
import org.springframework.web.bind.annotation.PostMapping;
import org.springframework.web.bind.annotation.RequestBody;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RequestParam;
import org.springframework.web.bind.annotation.ResponseStatus;
import org.springframework.web.bind.annotation.RestController;

@RestController
@RequestMapping("/api/favorites")
public class FavoritesController {
    private final FavoritesService favoritesService;

    public FavoritesController(FavoritesService favoritesService) {
        this.favoritesService = favoritesService;
    }

    @PostMapping
    @ResponseStatus(HttpStatus.CREATED)
    public void addFavorite(@Valid @RequestBody FavoriteRequest request) {
        favoritesService.addFavorite(request.getUserId(), request.getListingId());
    }

    @DeleteMapping("/{listingId}")
    @ResponseStatus(HttpStatus.NO_CONTENT)
    public void removeFavorite(@PathVariable String listingId, @RequestParam String userId) {
        favoritesService.removeFavorite(userId, listingId);
    }

    @GetMapping
    public List<String> listFavorites(@RequestParam String userId) {
        return favoritesService.listFavorites(userId);
    }
}
